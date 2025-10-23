package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "plan command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("plan command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
