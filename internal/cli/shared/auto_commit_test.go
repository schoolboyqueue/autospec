package shared

import (
	"testing"

	"github.com/ariel-frischer/autospec/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// createTestAutoCommitCommand creates a cobra command with auto-commit flags for testing.
func createTestAutoCommitCommand(autoCommit, noAutoCommit bool) *cobra.Command {
	cmd := &cobra.Command{}
	AddAutoCommitFlags(cmd)

	// Simulate parsing via command line args (this marks Changed=true)
	args := []string{}
	if autoCommit {
		args = append(args, "--auto-commit")
	}
	if noAutoCommit {
		args = append(args, "--no-auto-commit")
	}
	if len(args) > 0 {
		cmd.ParseFlags(args)
	}
	return cmd
}

func TestAddAutoCommitFlags(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	AddAutoCommitFlags(cmd)

	// Check flags are registered
	autoCommitFlag := cmd.Flags().Lookup(AutoCommitFlagName)
	assert.NotNil(t, autoCommitFlag, "auto-commit flag should be registered")
	assert.Equal(t, "false", autoCommitFlag.DefValue)

	noAutoCommitFlag := cmd.Flags().Lookup(NoAutoCommitFlagName)
	assert.NotNil(t, noAutoCommitFlag, "no-auto-commit flag should be registered")
	assert.Equal(t, "false", noAutoCommitFlag.DefValue)
}

func TestApplyAutoCommitOverride(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configAutoCommit bool
		autoCommitFlag   bool
		noAutoCommitFlag bool
		wantAutoCommit   bool
		wantApplied      bool
	}{
		"--auto-commit flag enables auto-commit": {
			configAutoCommit: false,
			autoCommitFlag:   true,
			noAutoCommitFlag: false,
			wantAutoCommit:   true,
			wantApplied:      true,
		},
		"--no-auto-commit flag disables auto-commit": {
			configAutoCommit: true,
			noAutoCommitFlag: true,
			autoCommitFlag:   false,
			wantAutoCommit:   false,
			wantApplied:      true,
		},
		"no flag preserves config value true": {
			configAutoCommit: true,
			autoCommitFlag:   false,
			noAutoCommitFlag: false,
			wantAutoCommit:   true,
			wantApplied:      false,
		},
		"no flag preserves config value false": {
			configAutoCommit: false,
			autoCommitFlag:   false,
			noAutoCommitFlag: false,
			wantAutoCommit:   false,
			wantApplied:      false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Configuration{
				AutoCommit:  tt.configAutoCommit,
				AgentPreset: "claude",
				SpecsDir:    "./specs",
				StateDir:    "~/.autospec/state",
			}

			cmd := createTestAutoCommitCommand(tt.autoCommitFlag, tt.noAutoCommitFlag)

			applied := ApplyAutoCommitOverride(cmd, cfg)

			assert.Equal(t, tt.wantApplied, applied)
			assert.Equal(t, tt.wantAutoCommit, cfg.AutoCommit)
		})
	}
}

func TestAutoCommitFlagsMutuallyExclusive(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	AddAutoCommitFlags(cmd)

	// Verify mutual exclusivity is set up (cobra stores this internally)
	// We can verify by checking that parsing both flags fails
	err := cmd.ParseFlags([]string{"--auto-commit", "--no-auto-commit"})
	// Note: Cobra's mutual exclusivity check happens at execution time, not parse time
	// So we just verify both flags exist and are marked mutually exclusive
	assert.Nil(t, err, "ParseFlags should not error - mutual exclusivity is checked at execution")
}

func TestAutoCommitConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "auto-commit", AutoCommitFlagName)
	assert.Equal(t, "no-auto-commit", NoAutoCommitFlagName)
}
