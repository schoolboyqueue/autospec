package cli

import (
	"fmt"
	"strings"

	"github.com/anthropics/auto-claude-speckit/internal/config"
	"github.com/anthropics/auto-claude-speckit/internal/workflow"
	"github.com/spf13/cobra"
)

var constitutionCmd = &cobra.Command{
	Use:   "constitution [optional-prompt]",
	Short: "Create or update the project constitution",
	Long: `Execute the /autospec.constitution command to create or update the project constitution.

The constitution command will:
- Create or update the project constitution in .specify/memory/constitution.md
- Define project principles and guidelines for development
- Can be run from any directory in the project

This command has no prerequisites - it can be run at any time.

You can optionally provide a prompt to guide the constitution generation:
  autospec constitution "Focus on security and performance"
  autospec constitution "Emphasize test-driven development"

Examples:
  autospec constitution                              # Generate/update constitution
  autospec constitution "Add API design guidelines"  # With specific guidance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get optional prompt from args
		var prompt string
		if len(args) > 0 {
			prompt = strings.Join(args, " ")
		}

		// Get flags
		configPath, _ := cmd.Flags().GetString("config")
		skipPreflight, _ := cmd.Flags().GetBool("skip-preflight")
		maxRetries, _ := cmd.Flags().GetInt("max-retries")

		// Load configuration
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override skip-preflight from flag if set
		if cmd.Flags().Changed("skip-preflight") {
			cfg.SkipPreflight = skipPreflight
		}

		// Override max-retries from flag if set
		if cmd.Flags().Changed("max-retries") {
			cfg.MaxRetries = maxRetries
		}

		// Create workflow orchestrator
		orch := workflow.NewWorkflowOrchestrator(cfg)

		// Execute constitution phase
		if err := orch.ExecuteConstitution(prompt); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(constitutionCmd)
}
