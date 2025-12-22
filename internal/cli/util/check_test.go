package util

import (
	"context"
	"net/http"
	"net/http/httptest"
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
		plain          bool
		responseCode   int
		responseBody   string
		wantErr        bool
		wantContains   []string
		wantNotContain []string
	}{
		"update available": {
			version:      "v0.6.0",
			plain:        false,
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
			plain:        false,
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
			plain:        false,
			responseCode: http.StatusOK,
			responseBody: `{}`,
			wantContains: []string{"dev", "update check"},
		},
		"rate limit error": {
			version:      "v0.6.0",
			plain:        false,
			responseCode: http.StatusForbidden,
			responseBody: `{"message": "rate limit exceeded"}`,
			wantContains: []string{"rate limit"},
		},
		"not found error": {
			version:      "v0.6.0",
			plain:        false,
			responseCode: http.StatusNotFound,
			responseBody: `{"message": "not found"}`,
			wantContains: []string{"No releases"},
		},
		"plain output update available": {
			version:      "v0.6.0",
			plain:        true,
			responseCode: http.StatusOK,
			responseBody: `{
				"tag_name": "v0.7.0",
				"published_at": "2025-12-20T00:00:00Z",
				"assets": [
					{"name": "autospec_0.7.0_Linux_x86_64.tar.gz", "browser_download_url": "https://example.com/download.tar.gz"},
					{"name": "checksums.txt", "browser_download_url": "https://example.com/checksums.txt"}
				]
			}`,
			wantContains: []string{"current:", "latest:", "update_available: true"},
		},
		"plain output up to date": {
			version:      "v0.7.0",
			plain:        true,
			responseCode: http.StatusOK,
			responseBody: `{
				"tag_name": "v0.7.0",
				"published_at": "2025-12-20T00:00:00Z",
				"assets": []
			}`,
			wantContains: []string{"current:", "latest:", "update_available: false"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Create mock server for GitHub API
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
				w.WriteHeader(tt.responseCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create a checker with mock API URL
			checker := update.NewChecker(5 * time.Second)
			checker.SetAPIURL(server.URL)

			// Execute check using the real implementation - pass plain as parameter
			output, err := executeCheck(context.Background(), checker, tt.version, tt.plain)

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

	// Create a slow server that times out
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name": "v0.7.0"}`))
	}))
	defer server.Close()

	// Create checker with very short timeout
	checker := update.NewChecker(10 * time.Millisecond)
	checker.SetAPIURL(server.URL)

	// Execute and verify timeout is handled - pass plain=false
	output, err := executeCheck(context.Background(), checker, "v0.6.0", false)

	// Should handle timeout gracefully with user-friendly message
	require.NoError(t, err)
	assert.Contains(t, output, "timeout", "output should mention timeout")
}

// TestRunCheck_ContextCancellation tests that the check command respects context cancellation.
func TestRunCheck_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name": "v0.7.0"}`))
	}))
	defer server.Close()

	checker := update.NewChecker(5 * time.Second)
	checker.SetAPIURL(server.URL)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Execute and verify cancellation is handled - pass plain=false
	_, err := executeCheck(ctx, checker, "v0.6.0", false)

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

	// Verify --plain flag is registered
	flag := ckCmd.Flags().Lookup("plain")
	assert.NotNil(t, flag, "--plain flag should be registered")
}

// TestCheckCommand_DevBuildMessage tests dev build specific messaging.
func TestCheckCommand_DevBuildMessage(t *testing.T) {
	t.Parallel()

	// For dev builds, we don't need a mock server since no network call should be made
	checker := update.NewChecker(5 * time.Second)
	output, err := executeCheck(context.Background(), checker, "dev", false)

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
		wantErr      bool
		wantContains string
	}{
		"invalid format": {
			version:      "invalid",
			wantErr:      true,
			wantContains: "version",
		},
		"partial version": {
			version:      "v1.0",
			wantErr:      true,
			wantContains: "version",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Create mock server - for invalid versions the check will fail during parsing
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"tag_name": "v1.0.0"}`))
			}))
			defer server.Close()

			checker := update.NewChecker(5 * time.Second)
			checker.SetAPIURL(server.URL)

			// Execute check - pass plain=false
			_, err := executeCheck(context.Background(), checker, tt.version, false)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantContains)
			}
		})
	}
}

// TestCheckCommand_PlainOutput tests the --plain flag output format.
func TestCheckCommand_PlainOutput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		version        string
		responseBody   string
		wantContains   []string
		wantNotContain []string
	}{
		"dev build plain": {
			version:      "dev",
			responseBody: `{}`,
			wantContains: []string{"version:", "status: dev-build"},
		},
		"error plain": {
			version:      "v0.6.0",
			responseBody: ``, // This will cause an error in parsing
			wantContains: []string{"error:"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.responseBody == "" {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			checker := update.NewChecker(5 * time.Second)
			checker.SetAPIURL(server.URL)

			// Pass plain=true for plain output
			output, err := executeCheck(context.Background(), checker, tt.version, true)

			// For plain output, errors should be formatted as output, not returned
			if err == nil {
				for _, want := range tt.wantContains {
					assert.Contains(t, output, want, "output should contain %q", want)
				}
			}
		})
	}
}

// createTestCommand creates a test cobra command for integration testing.
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.AddCommand(ckCmd)
	return cmd
}
