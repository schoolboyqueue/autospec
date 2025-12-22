package util

import (
	"context"
	"fmt"
	"strings"

	"github.com/ariel-frischer/autospec/internal/cli/shared"
	"github.com/ariel-frischer/autospec/internal/update"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var ckPlain bool

// ckCmd is the command for checking if an update is available.
var ckCmd = &cobra.Command{
	Use:     "ck",
	Aliases: []string{"check"},
	Short:   "Check if an update is available",
	Long:    "Check if a newer version of autospec is available on GitHub releases.",
	Example: `  # Check for available updates
  autospec ck

  # Plain output (for scripts)
  autospec ck --plain

  # Using the longer alias
  autospec check`,
	RunE: runCheck,
}

func init() {
	ckCmd.GroupID = shared.GroupGettingStarted
	ckCmd.Flags().BoolVar(&ckPlain, "plain", false, "Plain output without formatting")
}

// runCheck executes the update check command.
func runCheck(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	checker := update.NewChecker(update.DefaultHTTPTimeout)
	output, err := executeCheck(ctx, checker, Version, ckPlain)
	if err != nil {
		return err
	}

	fmt.Fprint(cmd.OutOrStdout(), output)
	return nil
}

// executeCheck performs the update check and returns formatted output.
// The plain parameter controls whether output is formatted for scripts.
func executeCheck(ctx context.Context, checker *update.Checker, version string, plain bool) (string, error) {
	// Handle dev builds without making network calls
	if version == "dev" || version == "" {
		return formatDevBuildMessage(version, plain), nil
	}

	// Perform the update check
	check, err := checker.CheckForUpdate(ctx, version)
	if err != nil {
		return handleCheckError(err, plain)
	}

	return formatCheckResult(check, plain), nil
}

// formatDevBuildMessage returns a message for dev builds.
func formatDevBuildMessage(version string, plain bool) string {
	if plain {
		return fmt.Sprintf("version: %s\nstatus: dev-build\nmessage: update check not applicable\n", version)
	}

	yellow := color.New(color.FgYellow).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()
	return fmt.Sprintf("%s Running dev build - update check not applicable\n%s\n",
		yellow("⚠"),
		dim("  Install a release version to enable update checks"))
}

// handleCheckError converts errors to user-friendly messages.
func handleCheckError(err error, plain bool) (string, error) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "rate limit"):
		return formatErrorMessage("GitHub API rate limit exceeded", "Please try again later", plain), nil
	case strings.Contains(errStr, "no releases"):
		return formatErrorMessage("No releases found", "No releases found on GitHub", plain), nil
	case strings.Contains(errStr, "deadline exceeded") || strings.Contains(errStr, "timeout"):
		return formatErrorMessage("Network timeout", "Network timeout while checking for updates", plain), nil
	case strings.Contains(errStr, "no asset found"):
		return formatErrorMessage("Platform not supported", "No release asset for this platform", plain), nil
	case strings.Contains(errStr, "context canceled"):
		return "", err
	default:
		return "", fmt.Errorf("checking for update: %w", err)
	}
}

// formatErrorMessage returns a formatted error message.
func formatErrorMessage(title, detail string, plain bool) string {
	if plain {
		return fmt.Sprintf("error: %s - %s\n", title, detail)
	}
	red := color.New(color.FgRed).SprintFunc()
	return fmt.Sprintf("%s %s\n  %s\n", red("✗"), title, detail)
}

// formatCheckResult formats the update check result for display.
func formatCheckResult(check *update.UpdateCheck, plain bool) string {
	if check.UpdateAvailable {
		return formatUpdateAvailable(check, plain)
	}
	return formatUpToDate(check, plain)
}

// formatUpdateAvailable returns output when an update is available.
func formatUpdateAvailable(check *update.UpdateCheck, plain bool) string {
	if plain {
		return fmt.Sprintf("current: %s\nlatest: %s\nupdate_available: true\n",
			check.CurrentVersion, check.LatestVersion)
	}

	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	return fmt.Sprintf("%s Update available: %s → %s\n%s\n",
		green("✓"),
		dim(check.CurrentVersion),
		cyan(check.LatestVersion),
		dim("  Run 'autospec update' to upgrade"))
}

// formatUpToDate returns output when already on latest version.
func formatUpToDate(check *update.UpdateCheck, plain bool) string {
	if plain {
		return fmt.Sprintf("current: %s\nlatest: %s\nupdate_available: false\n",
			check.CurrentVersion, check.LatestVersion)
	}

	green := color.New(color.FgGreen).SprintFunc()
	return fmt.Sprintf("%s Already on latest version (%s)\n",
		green("✓"),
		check.CurrentVersion)
}
