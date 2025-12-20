package cliagent

import (
	"context"
	"strings"
	"testing"
)

func TestNewCustomAgentFromConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config  CustomAgentConfig
		wantErr bool
		errMsg  string
	}{
		"valid config": {
			config: CustomAgentConfig{
				Command: "echo",
				Args:    []string{"{{PROMPT}}"},
			},
			wantErr: false,
		},
		"missing command": {
			config: CustomAgentConfig{
				Args: []string{"{{PROMPT}}"},
			},
			wantErr: true,
			errMsg:  "command is required",
		},
		"missing prompt placeholder": {
			config: CustomAgentConfig{
				Command: "echo",
				Args:    []string{"hello"},
			},
			wantErr: true,
			errMsg:  "must contain {{PROMPT}}",
		},
		"complex config": {
			config: CustomAgentConfig{
				Command: "aider",
				Args:    []string{"--model", "sonnet", "--yes-always", "--message", "{{PROMPT}}"},
			},
			wantErr: false,
		},
		"with env vars": {
			config: CustomAgentConfig{
				Command: "claude",
				Args:    []string{"-p", "{{PROMPT}}"},
				Env:     map[string]string{"ANTHROPIC_API_KEY": "test"},
			},
			wantErr: false,
		},
		"with post processor": {
			config: CustomAgentConfig{
				Command:       "claude",
				Args:          []string{"-p", "{{PROMPT}}"},
				PostProcessor: "cclean",
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := NewCustomAgentFromConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCustomAgentFromConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestCustomAgentConfig_IsValid(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config *CustomAgentConfig
		want   bool
	}{
		"nil config": {
			config: nil,
			want:   false,
		},
		"empty config": {
			config: &CustomAgentConfig{},
			want:   false,
		},
		"command only": {
			config: &CustomAgentConfig{Command: "echo"},
			want:   true,
		},
		"full config": {
			config: &CustomAgentConfig{
				Command: "claude",
				Args:    []string{"-p", "{{PROMPT}}"},
			},
			want: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := tt.config.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomAgent_Name(t *testing.T) {
	t.Parallel()
	agent, _ := NewCustomAgentFromConfig(CustomAgentConfig{
		Command: "echo",
		Args:    []string{"{{PROMPT}}"},
	})
	if got := agent.Name(); got != "custom" {
		t.Errorf("Name() = %q, want %q", got, "custom")
	}
}

func TestCustomAgent_Version(t *testing.T) {
	t.Parallel()
	agent, _ := NewCustomAgentFromConfig(CustomAgentConfig{
		Command: "echo",
		Args:    []string{"{{PROMPT}}"},
	})
	ver, err := agent.Version()
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	if ver != "custom" {
		t.Errorf("Version() = %q, want %q", ver, "custom")
	}
}

func TestCustomAgent_Capabilities(t *testing.T) {
	t.Parallel()
	agent, _ := NewCustomAgentFromConfig(CustomAgentConfig{
		Command: "echo",
		Args:    []string{"{{PROMPT}}"},
	})
	caps := agent.Capabilities()
	if !caps.Automatable {
		t.Error("Automatable should be true")
	}
	if caps.PromptDelivery.Method != PromptMethodTemplate {
		t.Errorf("PromptDelivery.Method = %q, want %q", caps.PromptDelivery.Method, PromptMethodTemplate)
	}
}

func TestCustomAgent_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config  CustomAgentConfig
		wantErr bool
		errMsg  string
	}{
		"valid command": {
			config: CustomAgentConfig{
				Command: "echo",
				Args:    []string{"{{PROMPT}}"},
			},
			wantErr: false,
		},
		"command not found": {
			config: CustomAgentConfig{
				Command: "nonexistent-cmd-12345",
				Args:    []string{"{{PROMPT}}"},
			},
			wantErr: true,
			errMsg:  "not found in PATH",
		},
		"post processor not found": {
			config: CustomAgentConfig{
				Command:       "echo",
				Args:          []string{"{{PROMPT}}"},
				PostProcessor: "nonexistent-processor-12345",
			},
			wantErr: true,
			errMsg:  "post_processor",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			agent, err := NewCustomAgentFromConfig(tt.config)
			if err != nil {
				t.Fatalf("NewCustomAgentFromConfig() error = %v", err)
			}
			err = agent.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestCustomAgent_BuildCommand(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config   CustomAgentConfig
		prompt   string
		wantCmd  string
		wantArgs []string
	}{
		"basic substitution": {
			config: CustomAgentConfig{
				Command: "echo",
				Args:    []string{"{{PROMPT}}"},
			},
			prompt:   "hello world",
			wantCmd:  "echo",
			wantArgs: []string{"hello world"},
		},
		"multiple args": {
			config: CustomAgentConfig{
				Command: "myapp",
				Args:    []string{"--message", "{{PROMPT}}", "--verbose"},
			},
			prompt:   "do something",
			wantCmd:  "myapp",
			wantArgs: []string{"--message", "do something", "--verbose"},
		},
		"prompt with special chars": {
			config: CustomAgentConfig{
				Command: "echo",
				Args:    []string{"{{PROMPT}}"},
			},
			prompt:   "hello \"world\" with 'quotes'",
			wantCmd:  "echo",
			wantArgs: []string{"hello \"world\" with 'quotes'"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			agent, _ := NewCustomAgentFromConfig(tt.config)
			cmd, err := agent.BuildCommand(tt.prompt, ExecOptions{})
			if err != nil {
				t.Fatalf("BuildCommand() error = %v", err)
			}
			if cmd.Path == "" {
				t.Error("cmd.Path should not be empty")
			}
			// For direct execution (no post-processor), Args[0] is the program path
			gotArgs := cmd.Args[1:]
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("args = %v, want %v", gotArgs, tt.wantArgs)
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

func TestCustomAgent_BuildCommand_WithPostProcessor(t *testing.T) {
	t.Parallel()

	agent, _ := NewCustomAgentFromConfig(CustomAgentConfig{
		Command:       "echo",
		Args:          []string{"{{PROMPT}}"},
		PostProcessor: "cat",
	})
	cmd, err := agent.BuildCommand("hello", ExecOptions{})
	if err != nil {
		t.Fatalf("BuildCommand() error = %v", err)
	}
	// Should use sh -c for piping
	if cmd.Args[0] != "sh" {
		t.Errorf("expected sh for shell execution, got %q", cmd.Args[0])
	}
	if cmd.Args[1] != "-c" {
		t.Errorf("expected -c flag, got %q", cmd.Args[1])
	}
	// The shell command should contain the pipe
	if !strings.Contains(cmd.Args[2], "|") {
		t.Errorf("expected pipe in shell command, got %q", cmd.Args[2])
	}
	if !strings.Contains(cmd.Args[2], "'cat'") {
		t.Errorf("expected post processor in shell command, got %q", cmd.Args[2])
	}
}

func TestCustomAgent_Execute(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config     CustomAgentConfig
		prompt     string
		wantStdout string
		wantExit   int
	}{
		"echo prompt": {
			config: CustomAgentConfig{
				Command: "echo",
				Args:    []string{"{{PROMPT}}"},
			},
			prompt:     "hello",
			wantStdout: "hello\n",
			wantExit:   0,
		},
		"exit code": {
			config: CustomAgentConfig{
				Command: "sh",
				Args:    []string{"-c", "{{PROMPT}}"},
			},
			prompt:   "exit 42",
			wantExit: 42,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			agent, _ := NewCustomAgentFromConfig(tt.config)
			result, err := agent.Execute(context.Background(), tt.prompt, ExecOptions{})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if result.ExitCode != tt.wantExit {
				t.Errorf("ExitCode = %d, want %d", result.ExitCode, tt.wantExit)
			}
			if tt.wantStdout != "" && result.Stdout != tt.wantStdout {
				t.Errorf("Stdout = %q, want %q", result.Stdout, tt.wantStdout)
			}
		})
	}
}

func TestCustomAgent_PromptWithNewlines(t *testing.T) {
	t.Parallel()
	agent, _ := NewCustomAgentFromConfig(CustomAgentConfig{
		Command: "echo",
		Args:    []string{"{{PROMPT}}"},
	})
	prompt := "line1\nline2\nline3"
	cmd, err := agent.BuildCommand(prompt, ExecOptions{})
	if err != nil {
		t.Fatalf("BuildCommand() error = %v", err)
	}
	// The prompt should be preserved with newlines
	if len(cmd.Args) < 2 {
		t.Fatal("expected at least 2 args")
	}
	if !strings.Contains(cmd.Args[1], "\n") {
		t.Errorf("prompt should contain newlines, got %q", cmd.Args[1])
	}
}

func TestCustomAgent_WorkDir(t *testing.T) {
	t.Parallel()
	agent, _ := NewCustomAgentFromConfig(CustomAgentConfig{
		Command: "pwd",
		Args:    []string{"{{PROMPT}}"},
	})
	cmd, _ := agent.BuildCommand("ignored", ExecOptions{WorkDir: "/tmp"})
	if cmd.Dir != "/tmp" {
		t.Errorf("cmd.Dir = %q, want %q", cmd.Dir, "/tmp")
	}
}

func TestCustomAgent_Env(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config   CustomAgentConfig
		opts     ExecOptions
		wantEnvs []string
	}{
		"config env vars": {
			config: CustomAgentConfig{
				Command: "echo",
				Args:    []string{"{{PROMPT}}"},
				Env:     map[string]string{"CONFIG_VAR": "config_value"},
			},
			opts:     ExecOptions{},
			wantEnvs: []string{"CONFIG_VAR=config_value"},
		},
		"exec env vars": {
			config: CustomAgentConfig{
				Command: "echo",
				Args:    []string{"{{PROMPT}}"},
			},
			opts:     ExecOptions{Env: map[string]string{"EXEC_VAR": "exec_value"}},
			wantEnvs: []string{"EXEC_VAR=exec_value"},
		},
		"both env vars": {
			config: CustomAgentConfig{
				Command: "echo",
				Args:    []string{"{{PROMPT}}"},
				Env:     map[string]string{"CONFIG_VAR": "config_value"},
			},
			opts:     ExecOptions{Env: map[string]string{"EXEC_VAR": "exec_value"}},
			wantEnvs: []string{"CONFIG_VAR=config_value", "EXEC_VAR=exec_value"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			agent, _ := NewCustomAgentFromConfig(tt.config)
			cmd, _ := agent.BuildCommand("test", tt.opts)

			for _, wantEnv := range tt.wantEnvs {
				found := false
				for _, e := range cmd.Env {
					if e == wantEnv {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %s in env", wantEnv)
				}
			}
		})
	}
}
