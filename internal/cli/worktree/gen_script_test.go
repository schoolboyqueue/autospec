package worktree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenScriptCmd_Structure(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "gen-script", genScriptCmd.Use)
	assert.NotEmpty(t, genScriptCmd.Short)
	assert.NotEmpty(t, genScriptCmd.Long)
}

func TestGenScriptCmd_Flags(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		flagName  string
		shorthand string
		wantExist bool
	}{
		"include-env flag exists": {
			flagName:  "include-env",
			shorthand: "",
			wantExist: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			flag := genScriptCmd.Flags().Lookup(tt.flagName)
			if tt.wantExist {
				assert.NotNil(t, flag, "flag %q should exist", tt.flagName)
				if tt.shorthand != "" {
					assert.Equal(t, tt.shorthand, flag.Shorthand)
				}
			} else {
				assert.Nil(t, flag, "flag %q should not exist", tt.flagName)
			}
		})
	}
}

func TestGenScriptCmd_RegisteredUnderWorktree(t *testing.T) {
	// Not parallel: WorktreeCmd.Commands() has lazy init that races with other tests
	subcommands := WorktreeCmd.Commands()
	names := make([]string, len(subcommands))
	for i, cmd := range subcommands {
		names[i] = cmd.Name()
	}

	assert.Contains(t, names, "gen-script", "gen-script should be registered under worktree command")
}

func TestBuildWorktreeSetupCommand(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		includeEnv bool
		want       string
	}{
		"without include-env": {
			includeEnv: false,
			want:       "/autospec.worktree-setup",
		},
		"with include-env": {
			includeEnv: true,
			want:       "/autospec.worktree-setup --include-env",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := buildWorktreeSetupCommand(tt.includeEnv)
			assert.Equal(t, tt.want, got)
		})
	}
}
