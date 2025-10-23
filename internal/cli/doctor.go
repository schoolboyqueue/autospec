package cli

import (
	"fmt"
	"os"

	"github.com/anthropics/auto-claude-speckit/internal/health"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run health checks for autospec dependencies",
	Long: `Run health checks to verify that all required dependencies are installed and available.

This command checks for:
  - Claude CLI
  - Specify CLI
  - Git

Each check will display a ✓ if passed or ✗ with an error message if failed.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Run all health checks
		report := health.RunHealthChecks()

		// Format and display the report
		output := health.FormatReport(report)
		fmt.Print(output)

		// Exit with non-zero status if any checks failed
		if !report.Passed {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
