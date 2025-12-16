package cli

import (
	"fmt"
	"strings"

	"github.com/ariel-frischer/autospec/internal/completion"
	"github.com/spf13/cobra"
)

var completionInstallCmd = &cobra.Command{
	Use:   "install [bash|zsh|fish|powershell]",
	Short: "Install shell completions for autospec",
	Long: `Install shell completions for autospec.

This command auto-detects your shell from the $SHELL environment variable
and installs completions appropriately:

  - Bash: Appends sourcing block to ~/.bashrc
  - Zsh: Appends sourcing block to ~/.zshrc
  - Fish: Writes completion file to ~/.config/fish/completions/
  - PowerShell: Appends sourcing block to $PROFILE

A backup is created before modifying any rc file (with .autospec-backup-TIMESTAMP suffix).

Use the --manual flag to display manual installation instructions without
modifying any files.`,
	Example: `  # Auto-detect shell and install
  autospec completion install

  # Install for a specific shell
  autospec completion install bash
  autospec completion install zsh
  autospec completion install fish
  autospec completion install powershell

  # Show manual installation instructions
  autospec completion install --manual
  autospec completion install bash --manual`,
	Args: cobra.MaximumNArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: runCompletionInstall,
}

var manualFlag bool

func init() {
	// Find and add to the completion command
	// Cobra generates the completion command automatically, so we need to
	// add our subcommand after rootCmd is fully initialized
	cobra.OnInitialize(func() {
		// Find the completion command
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "completion" {
				cmd.AddCommand(completionInstallCmd)
				break
			}
		}
	})

	completionInstallCmd.Flags().BoolVar(&manualFlag, "manual", false, "Show manual installation instructions without modifying files")
}

func runCompletionInstall(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	// Determine the shell
	var shell completion.Shell
	var err error

	if len(args) > 0 {
		// Explicit shell argument
		shellArg := strings.ToLower(args[0])
		if !completion.IsValidShell(shellArg) {
			supportedShells := make([]string, len(completion.SupportedShells()))
			for i, s := range completion.SupportedShells() {
				supportedShells[i] = string(s)
			}
			return fmt.Errorf("unknown shell: %s\nSupported shells: %s", shellArg, strings.Join(supportedShells, ", "))
		}
		shell = completion.Shell(shellArg)
	} else {
		// Auto-detect shell
		shell, err = completion.DetectShell()
		if err != nil {
			// If detection fails, show help with all manual instructions
			fmt.Fprintln(out, "Could not auto-detect shell:", err)
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Please specify a shell explicitly:")
			fmt.Fprintln(out, "  autospec completion install bash")
			fmt.Fprintln(out, "  autospec completion install zsh")
			fmt.Fprintln(out, "  autospec completion install fish")
			fmt.Fprintln(out, "  autospec completion install powershell")
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Or use --manual for installation instructions:")
			fmt.Fprintln(out, "  autospec completion install --manual")
			return nil
		}
		fmt.Fprintf(out, "Detected shell: %s\n", shell)
	}

	// Handle --manual flag
	if manualFlag {
		fmt.Fprintln(out, completion.GetManualInstructions(shell))
		return nil
	}

	// Perform installation
	result, err := completion.Install(shell)
	if err != nil {
		// Check for permission error
		if completion.IsPermissionError(err) {
			fmt.Fprintf(out, "Error: %v\n\n", err)
			fmt.Fprintln(out, "Automatic installation failed. Here are manual instructions:")
			fmt.Fprintln(out)
			fmt.Fprintln(out, completion.GetManualInstructions(shell))
			return nil
		}
		return err
	}

	// Display result
	fmt.Fprintln(out, result.Message)

	return nil
}
