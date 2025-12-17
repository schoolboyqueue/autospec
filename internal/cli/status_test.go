package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusCmdRegistration(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "status [spec-name]" {
			found = true
			break
		}
	}
	assert.True(t, found, "status command should be registered")
}

func TestStatusCmdFlags(t *testing.T) {
	// verbose flag
	f := statusCmd.Flags().Lookup("verbose")
	require.NotNil(t, f)
	assert.Equal(t, "v", f.Shorthand)
	assert.Equal(t, "false", f.DefValue)
}

func TestStatusCmdArgs(t *testing.T) {
	// Should accept 0 or 1 args
	err := statusCmd.Args(statusCmd, []string{})
	assert.NoError(t, err)

	err = statusCmd.Args(statusCmd, []string{"spec-name"})
	assert.NoError(t, err)

	err = statusCmd.Args(statusCmd, []string{"arg1", "arg2"})
	assert.Error(t, err)
}

func TestStatusCmdAlias(t *testing.T) {
	assert.Contains(t, statusCmd.Aliases, "st", "status command should have 'st' alias")
}

func TestStatusCmdSilenceUsage(t *testing.T) {
	assert.True(t, statusCmd.SilenceUsage, "status command should silence usage on errors")
}

func TestStatusCmdDefaultVerbose(t *testing.T) {
	// Default verbose should be false
	verbose, _ := statusCmd.Flags().GetBool("verbose")
	assert.False(t, verbose)
}
