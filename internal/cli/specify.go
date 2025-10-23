package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

var specifyCmd = &cobra.Command{
	Use:   "specify",
	Short: "specify command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("specify command not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(specifyCmd)
}
