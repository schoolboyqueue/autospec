package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/auto-claude-speckit/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize autospec configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		global, _ := cmd.Flags().GetBool("global")
		force, _ := cmd.Flags().GetBool("force")

		configPath := ".autospec/config.json"
		if global {
			homeDir, _ := os.UserHomeDir()
			configPath = filepath.Join(homeDir, ".autospec", "config.json")
		}

		if _, err := os.Stat(configPath); err == nil && !force {
			return fmt.Errorf("config exists at %s (use --force)", configPath)
		}

		os.MkdirAll(filepath.Dir(configPath), 0755)
		defaults := config.GetDefaults()
		data, _ := json.MarshalIndent(defaults, "", "  ")
		os.WriteFile(configPath, data, 0644)

		fmt.Printf("Created configuration at %s\n", configPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("global", "g", false, "Create global config")
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing")
}
