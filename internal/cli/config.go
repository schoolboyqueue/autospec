package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("config command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
