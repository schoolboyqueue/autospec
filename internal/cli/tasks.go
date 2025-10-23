package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "tasks command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("tasks command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(tasksCmd)
}
