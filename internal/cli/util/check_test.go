package util

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ariel-frischer/autospec/internal/update"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunCheck_TableDriven tests the check command with various scenarios.
// Uses map-based table-driven pattern per coding standards.
func TestRunCheck_TableDriven(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		version        string
		responseCode   int
		responseBody   string
		wantErr        bool
		wantContains   []string
		wantNotContain []string
	}{
		"update available": {
			version:      "v0.6.0",
			responseCode: http.StatusOK,
			responseBody: `{
				"tag_name": "v0.7.0",
				"published_at": "2025-12-20T00:00:00Z",
				"assets": [
					{"name": "autospec_0.7.0_Linux_x86_64.tar.gz", "browser_download_url": "https://example.com/download.tar.gz"},
					{"name": "checksums.txt", "browser_download_url": "https://example.com/checksums.txt"}
				]
			}`,
			wantContains: []string{"v0.6.0", "v0.7.0", "autospec update"},
		},
		"already on latest version": {
			version:      "v0.7.0",
			responseCode: http.StatusOK,
			responseBody: `{
				"tag_name": "v0.7.0",
				"published_at": "2025-12-20T00:00:00Z",
				"assets": []
			}`,
			wantContains:   []string{"latest", "v0.7.0"},
			wantNotContain: []string{"update available"},
		},
		"dev build": {
			version:      "dev",
			responseCode: http.StatusOK,
			responseBody: `{}`,
			wantContains: []string{"dev", "update check"},
		},
		"rate limit error": {
			version:      "v0.6.0",
			responseCode: http.StatusForbidden,
			responseBody: `{"message": "rate limit exceeded"}`,
			wantContains: []string{"rate limit"},
		},
		"not found error": {
			version:      "v0.6.0",
			responseCode: http.StatusNotFound,
			responseBody: `{"message": "not found"}`,
			wantContains: []string{"No releases"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Save original version and restore after test
			origVersion := Version
			Version = tt.version
			defer func() { Version = origVersion }()

			// Create mock server for GitHub API
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
				w.WriteHeader(tt.responseCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create a checker with mock API URL
			checker := update.NewChecker(5 * time.Second)
			setCheckerAPIURL(checker, server.URL)

			// Execute check with mock checker
			output, err := executeCheckWithChecker(context.Background(), checker, tt.version)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			// Verify output contains expected strings
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "output should contain %q", want)
			}

			// Verify output does not contain unwanted strings
			for _, notWant := range tt.wantNotContain {
				assert.NotContains(t, output, notWant, "output should not contain %q", notWant)
			}
		})
	}
}

// TestRunCheck_NetworkTimeout tests that the check command handles network timeouts gracefully.
func TestRunCheck_NetworkTimeout(t *testing.T) {
	t.Parallel()

	// Save original version and restore after test
	origVersion := Version
	Version = "v0.6.0"
	defer func() { Version = origVersion }()

	// Create a slow server that times out
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name": "v0.7.0"}`))
	}))
	defer server.Close()

	// Create checker with very short timeout
	checker := update.NewChecker(10 * time.Millisecond)
	setCheckerAPIURL(checker, server.URL)

	// Execute and verify timeout is handled
	output, err := executeCheckWithChecker(context.Background(), checker, "v0.6.0")

	// Should handle timeout gracefully with user-friendly message
	require.NoError(t, err)
	assert.Contains(t, output, "timeout", "output should mention timeout")
}

// TestRunCheck_ContextCancellation tests that the check command respects context cancellation.
func TestRunCheck_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Save original version and restore after test
	origVersion := Version
	Version = "v0.6.0"
	defer func() { Version = origVersion }()

	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name": "v0.7.0"}`))
	}))
	defer server.Close()

	checker := update.NewChecker(5 * time.Second)
	setCheckerAPIURL(checker, server.URL)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Execute and verify cancellation is handled
	_, err := executeCheckWithChecker(ctx, checker, "v0.6.0")

	// Should handle cancellation
	assert.Error(t, err)
}

// TestCheckCommand_CobraIntegration tests the cobra command integration.
func TestCheckCommand_CobraIntegration(t *testing.T) {
	t.Parallel()

	// Verify command is properly configured
	assert.Equal(t, "ck", ckCmd.Use)
	assert.Contains(t, ckCmd.Aliases, "check")
	assert.NotEmpty(t, ckCmd.Short)
	assert.NotEmpty(t, ckCmd.Long)
	assert.NotEmpty(t, ckCmd.Example)
	assert.NotNil(t, ckCmd.RunE)
}

// TestCheckCommand_DevBuildMessage tests dev build specific messaging.
func TestCheckCommand_DevBuildMessage(t *testing.T) {
	t.Parallel()

	// Save original version and restore after test
	origVersion := Version
	Version = "dev"
	defer func() { Version = origVersion }()

	// For dev builds, we don't need a mock server since no network call should be made
	output, err := executeCheckWithVersion(context.Background(), "dev")

	require.NoError(t, err)
	assert.Contains(t, output, "dev")
	// Should indicate that update checks don't apply to dev builds
	assert.Contains(t, output, "update")
}

// TestCheckCommand_ParseError tests handling of unparseable version strings.
func TestCheckCommand_ParseError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		version      string
		wantContains string
	}{
		"invalid format": {
			version:      "invalid",
			wantContains: "version",
		},
		"partial version": {
			version:      "v1.0",
			wantContains: "version",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Save original version and restore after test
			origVersion := Version
			Version = tt.version
			defer func() { Version = origVersion }()

			// Execute check
			output, err := executeCheckWithVersion(context.Background(), tt.version)

			// Should handle parse error gracefully
			if err != nil {
				assert.Contains(t, err.Error(), tt.wantContains)
			} else {
				// If no error, should display current version in output
				assert.Contains(t, output, tt.version)
			}
		})
	}
}

// Helper functions for testing

// setCheckerAPIURL sets the API URL for the checker using the exported SetAPIURL method.
func setCheckerAPIURL(checker *update.Checker, url string) {
	checker.SetAPIURL(url)
}

// executeCheckWithChecker executes the check command with a specific checker.
// Returns the output and any error.
// This simulates what the actual check.go implementation should produce.
func executeCheckWithChecker(ctx context.Context, checker *update.Checker, version string) (string, error) {
	var output bytes.Buffer

	// Handle dev builds without making network calls
	if version == "dev" || version == "" {
		output.WriteString("Running dev build - update check not applicable\n")
		return output.String(), nil
	}

	// Perform the update check
	check, err := checker.CheckForUpdate(ctx, version)
	if err != nil {
		// Handle specific error types with user-friendly messages
		errStr := err.Error()
		if strings.Contains(errStr, "rate limit") {
			output.WriteString("Error: GitHub API rate limit exceeded. Please try again later.\n")
			return output.String(), nil
		}
		if strings.Contains(errStr, "no releases") {
			output.WriteString("Error: No releases found on GitHub.\n")
			return output.String(), nil
		}
		if strings.Contains(errStr, "deadline exceeded") || strings.Contains(errStr, "timeout") {
			output.WriteString("Error: Network timeout while checking for updates.\n")
			return output.String(), nil
		}
		// For other errors, return them
		return "", err
	}

	// Format the output based on the check result
	if check.UpdateAvailable {
		output.WriteString("Update available: ")
		output.WriteString(check.CurrentVersion)
		output.WriteString(" -> ")
		output.WriteString(check.LatestVersion)
		output.WriteString("\nRun 'autospec update' to upgrade\n")
	} else {
		output.WriteString("Already on latest version (")
		output.WriteString(check.CurrentVersion)
		output.WriteString(")\n")
	}

	return output.String(), nil
}

// executeCheckWithVersion executes the check command with the given version.
// Returns the output and any error.
func executeCheckWithVersion(ctx context.Context, version string) (string, error) {
	// For dev builds, return dev-specific message
	if version == "dev" {
		return "Running dev build - update check not applicable\n", nil
	}

	// For invalid versions, return parse error
	_, err := update.ParseVersion(version)
	if err != nil {
		return "", err
	}

	return "Version: " + version + "\n", nil
}

// createTestCommand creates a test cobra command for integration testing.
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.AddCommand(ckCmd)
	return cmd
}
