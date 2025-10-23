package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

var implementCmd = &cobra.Command{
	Use:   "implement",
	Short: "implement command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("implement command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(implementCmd)
}
