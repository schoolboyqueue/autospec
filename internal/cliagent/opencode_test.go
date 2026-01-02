package cliagent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ariel-frischer/autospec/internal/opencode"
)

func TestNewOpenCode(t *testing.T) {
	t.Parallel()
	agent := NewOpenCode()

	t.Run("agent name", func(t *testing.T) {
		t.Parallel()
		if got := agent.Name(); got != "opencode" {
			t.Errorf("Name() = %q, want %q", got, "opencode")
		}
	})

	t.Run("command", func(t *testing.T) {
		t.Parallel()
		if agent.Cmd != "opencode" {
			t.Errorf("Cmd = %q, want %q", agent.Cmd, "opencode")
		}
	})

	t.Run("version flag", func(t *testing.T) {
		t.Parallel()
		if agent.VersionFlag != "--version" {
			t.Errorf("VersionFlag = %q, want %q", agent.VersionFlag, "--version")
		}
	})

	t.Run("prompt delivery method", func(t *testing.T) {
		t.Parallel()
		caps := agent.Capabilities()
		if caps.PromptDelivery.Method != PromptMethodSubcommandWithFlag {
			t.Errorf("PromptDelivery.Method = %q, want %q",
				caps.PromptDelivery.Method, PromptMethodSubcommandWithFlag)
		}
	})

	t.Run("prompt delivery flag", func(t *testing.T) {
		t.Parallel()
		caps := agent.Capabilities()
		if caps.PromptDelivery.Flag != "run" {
			t.Errorf("PromptDelivery.Flag = %q, want %q",
				caps.PromptDelivery.Flag, "run")
		}
	})

	t.Run("prompt delivery command flag", func(t *testing.T) {
		t.Parallel()
		caps := agent.Capabilities()
		if caps.PromptDelivery.CommandFlag != "--command" {
			t.Errorf("PromptDelivery.CommandFlag = %q, want %q",
				caps.PromptDelivery.CommandFlag, "--command")
		}
	})

	t.Run("automatable", func(t *testing.T) {
		t.Parallel()
		caps := agent.Capabilities()
		if !caps.Automatable {
			t.Error("Automatable should be true")
		}
	})

	t.Run("no autonomous flag", func(t *testing.T) {
		t.Parallel()
		caps := agent.Capabilities()
		if caps.AutonomousFlag != "" {
			t.Errorf("AutonomousFlag = %q, want empty (run subcommand is inherently non-interactive)",
				caps.AutonomousFlag)
		}
	})
}

func TestOpenCode_BuildCommand(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		prompt   string
		opts     ExecOptions
		wantArgs []string
	}{
		"basic prompt": {
			prompt:   "fix the bug",
			opts:     ExecOptions{},
			wantArgs: []string{"run", "fix the bug"},
		},
		"with slash command via ExtraArgs": {
			prompt:   "specify this feature",
			opts:     ExecOptions{ExtraArgs: []string{"--command", "autospec.specify"}},
			wantArgs: []string{"run", "specify this feature", "--command", "autospec.specify"},
		},
		"with plan command via ExtraArgs": {
			prompt:   "plan the implementation",
			opts:     ExecOptions{ExtraArgs: []string{"--command", "autospec.plan"}},
			wantArgs: []string{"run", "plan the implementation", "--command", "autospec.plan"},
		},
		"with implement command via ExtraArgs": {
			prompt:   "implement the feature",
			opts:     ExecOptions{ExtraArgs: []string{"--command", "autospec.implement"}},
			wantArgs: []string{"run", "implement the feature", "--command", "autospec.implement"},
		},
		"with retry prompt injection": {
			prompt: `Original task: implement feature

## Validation Errors (Retry 1/3)
- Missing required field: 'id'
- Invalid format for field 'description'

Please fix these validation errors and try again.`,
			opts: ExecOptions{ExtraArgs: []string{"--command", "autospec.specify"}},
			wantArgs: []string{"run", `Original task: implement feature

## Validation Errors (Retry 1/3)
- Missing required field: 'id'
- Invalid format for field 'description'

Please fix these validation errors and try again.`, "--command", "autospec.specify"},
		},
		"with multiple extra args": {
			prompt:   "analyze the code",
			opts:     ExecOptions{ExtraArgs: []string{"--model", "opus", "--command", "autospec.analyze"}},
			wantArgs: []string{"run", "analyze the code", "--model", "opus", "--command", "autospec.analyze"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			agent := NewOpenCode()
			cmd, err := agent.BuildCommand(tt.prompt, tt.opts)
			if err != nil {
				t.Fatalf("BuildCommand() error = %v", err)
			}
			if len(cmd.Args) < 1 {
				t.Fatal("BuildCommand() returned cmd with no args")
			}
			gotArgs := cmd.Args[1:] // Skip the command name ("opencode")
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("args len = %d, want %d\ngot: %v\nwant: %v",
					len(gotArgs), len(tt.wantArgs), gotArgs, tt.wantArgs)
				return
			}
			for i, arg := range gotArgs {
				if arg != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestOpenCode_BuildCommand_CommandName(t *testing.T) {
	t.Parallel()
	agent := NewOpenCode()

	// Verify the command executable name
	cmd, err := agent.BuildCommand("test prompt", ExecOptions{})
	if err != nil {
		t.Fatalf("BuildCommand() error = %v", err)
	}
	if cmd.Args[0] != "opencode" {
		t.Errorf("command = %q, want %q", cmd.Args[0], "opencode")
	}
}

func TestOpenCode_BuildCommand_Pattern(t *testing.T) {
	t.Parallel()
	agent := NewOpenCode()

	// Test the expected pattern: opencode run <prompt> --command <command-name>
	opts := ExecOptions{ExtraArgs: []string{"--command", "autospec.specify"}}
	cmd, err := agent.BuildCommand("specify my feature", opts)
	if err != nil {
		t.Fatalf("BuildCommand() error = %v", err)
	}

	args := cmd.Args
	// Expected: ["opencode", "run", "specify my feature", "--command", "autospec.specify"]
	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d: %v", len(args), args)
	}
	if args[0] != "opencode" {
		t.Errorf("args[0] = %q, want %q", args[0], "opencode")
	}
	if args[1] != "run" {
		t.Errorf("args[1] = %q, want %q", args[1], "run")
	}
	if args[2] != "specify my feature" {
		t.Errorf("args[2] = %q, want %q", args[2], "specify my feature")
	}
	if args[3] != "--command" {
		t.Errorf("args[3] = %q, want %q", args[3], "--command")
	}
	if args[4] != "autospec.specify" {
		t.Errorf("args[4] = %q, want %q", args[4], "autospec.specify")
	}
}
