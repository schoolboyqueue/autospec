package cli

import (
	"strings"
	"testing"
)

func TestCommandAliases(t *testing.T) {
	tests := []struct {
		name          string
		commandName   string
		expectedAlias []string
	}{
		{
			name:          "specify command has aliases spec and s",
			commandName:   "specify",
			expectedAlias: []string{"spec", "s"},
		},
		{
			name:          "plan command has alias p",
			commandName:   "plan",
			expectedAlias: []string{"p"},
		},
		{
			name:          "tasks command has alias t",
			commandName:   "tasks",
			expectedAlias: []string{"t"},
		},
		{
			name:          "implement command has aliases impl and i",
			commandName:   "implement",
			expectedAlias: []string{"impl", "i"},
		},
		{
			name:          "status command has alias st",
			commandName:   "status",
			expectedAlias: []string{"st"},
		},
		{
			name:          "doctor command has alias doc",
			commandName:   "doctor",
			expectedAlias: []string{"doc"},
		},
		{
			name:          "constitution command has alias const",
			commandName:   "constitution",
			expectedAlias: []string{"const"},
		},
		{
			name:          "clarify command has alias cl",
			commandName:   "clarify",
			expectedAlias: []string{"cl"},
		},
		{
			name:          "checklist command has alias chk",
			commandName:   "checklist",
			expectedAlias: []string{"chk"},
		},
		{
			name:          "analyze command has alias az",
			commandName:   "analyze",
			expectedAlias: []string{"az"},
		},
		{
			name:          "version command has alias v",
			commandName:   "version",
			expectedAlias: []string{"v"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find the command by name
			cmd, _, err := rootCmd.Find([]string{tt.commandName})
			if err != nil {
				t.Fatalf("command %q not found: %v", tt.commandName, err)
			}

			// Verify aliases
			if len(cmd.Aliases) != len(tt.expectedAlias) {
				t.Errorf("command %q has %d aliases, want %d; got %v, want %v",
					tt.commandName, len(cmd.Aliases), len(tt.expectedAlias), cmd.Aliases, tt.expectedAlias)
				return
			}

			for i, expected := range tt.expectedAlias {
				if cmd.Aliases[i] != expected {
					t.Errorf("command %q alias[%d] = %q, want %q", tt.commandName, i, cmd.Aliases[i], expected)
				}
			}
		})
	}
}

func TestAliasResolution(t *testing.T) {
	tests := []struct {
		alias       string
		commandName string
	}{
		{"spec", "specify"},
		{"s", "specify"},
		{"p", "plan"},
		{"t", "tasks"},
		{"impl", "implement"},
		{"i", "implement"},
		{"st", "status"},
		{"doc", "doctor"},
		{"const", "constitution"},
		{"cl", "clarify"},
		{"chk", "checklist"},
		{"az", "analyze"},
		{"v", "version"},
	}

	for _, tt := range tests {
		t.Run(tt.alias+" resolves to "+tt.commandName, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{tt.alias})
			if err != nil {
				t.Fatalf("alias %q not found: %v", tt.alias, err)
			}

			if cmd.Name() != tt.commandName {
				t.Errorf("alias %q resolved to %q, want %q", tt.alias, cmd.Name(), tt.commandName)
			}
		})
	}
}

func TestHelpTextShowsAliases(t *testing.T) {
	tests := []struct {
		commandName    string
		expectedInHelp string
	}{
		{"specify", "Aliases:"},
		{"plan", "Aliases:"},
		{"tasks", "Aliases:"},
		{"implement", "Aliases:"},
		{"status", "Aliases:"},
		{"doctor", "Aliases:"},
		{"constitution", "Aliases:"},
		{"clarify", "Aliases:"},
		{"checklist", "Aliases:"},
		{"analyze", "Aliases:"},
		{"version", "Aliases:"},
	}

	for _, tt := range tests {
		t.Run(tt.commandName+" help shows aliases", func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{tt.commandName})
			if err != nil {
				t.Fatalf("command %q not found: %v", tt.commandName, err)
			}

			// Get the usage string which includes aliases
			usage := cmd.UsageString()
			if !strings.Contains(usage, tt.expectedInHelp) {
				t.Errorf("command %q help does not contain %q:\n%s", tt.commandName, tt.expectedInHelp, usage)
			}
		})
	}
}
