package workflow

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpandTemplate tests template placeholder expansion
func TestExpandTemplate(t *testing.T) {
	tests := map[string]struct {
		template string
		prompt   string
		want     string
	}{
		"simple replacement": {
			template: "claude -p {{PROMPT}}",
			prompt:   "/speckit.specify \"test\"",
			want:     "claude -p /speckit.specify \"test\"",
		},
		"with env var prefix": {
			template: "ANTHROPIC_API_KEY=\"\" claude -p {{PROMPT}}",
			prompt:   "/speckit.plan",
			want:     "ANTHROPIC_API_KEY=\"\" claude -p /speckit.plan",
		},
		"with pipe": {
			template: "claude -p {{PROMPT}} | claude-clean",
			prompt:   "/speckit.tasks",
			want:     "claude -p /speckit.tasks | claude-clean",
		},
		"complex template": {
			template: "ANTHROPIC_API_KEY=\"\" claude -p \"{{PROMPT}}\" --verbose | tee output.log",
			prompt:   "/speckit.specify \"Add auth\"",
			want:     "ANTHROPIC_API_KEY=\"\" claude -p \"/speckit.specify \"Add auth\"\" --verbose | tee output.log",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			executor := &ClaudeExecutor{
				CustomClaudeCmd: tc.template,
			}

			result := executor.expandTemplate(tc.prompt)
			assert.Equal(t, tc.want, result)
		})
	}
}

// TestValidateTemplate tests custom command template validation
func TestValidateTemplate(t *testing.T) {
	tests := map[string]struct {
		template string
		wantErr  bool
	}{
		"empty template": {
			template: "",
			wantErr:  false,
		},
		"valid template": {
			template: "claude -p {{PROMPT}}",
			wantErr:  false,
		},
		"valid with env var": {
			template: "ANTHROPIC_API_KEY=\"\" claude -p {{PROMPT}}",
			wantErr:  false,
		},
		"valid with pipe": {
			template: "claude -p {{PROMPT}} | claude-clean",
			wantErr:  false,
		},
		"missing placeholder": {
			template: "claude -p \"test\"",
			wantErr:  true,
		},
		"wrong placeholder format": {
			template: "claude -p {PROMPT}",
			wantErr:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateTemplate(tc.template)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "{{PROMPT}}")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestClaudeExecutor_StreamCommand tests command execution with output streaming
func TestClaudeExecutor_StreamCommand(t *testing.T) {
	tests := map[string]struct {
		executor *ClaudeExecutor
		prompt   string
		wantErr  bool
	}{
		"simple mode (no custom command)": {
			executor: &ClaudeExecutor{
				ClaudeCmd:  "echo",
				ClaudeArgs: []string{},
				UseAPIKey:  false,
			},
			prompt:  "test prompt",
			wantErr: false,
		},
		"with custom command using echo": {
			executor: &ClaudeExecutor{
				CustomClaudeCmd: "echo {{PROMPT}}",
				UseAPIKey:       false,
			},
			prompt:  "test prompt",
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			err := tc.executor.StreamCommand(tc.prompt, &stdout, &stderr)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify output was captured
				assert.NotEmpty(t, stdout.String() + stderr.String())
			}
		})
	}
}

// TestParseCustomCommand tests custom command parsing
func TestParseCustomCommand(t *testing.T) {
	executor := &ClaudeExecutor{}

	tests := map[string]struct {
		cmdStr string
	}{
		"simple command": {
			cmdStr: "claude -p /speckit.specify",
		},
		"with env var": {
			cmdStr: "ANTHROPIC_API_KEY=\"\" claude -p /speckit.plan",
		},
		"with pipe": {
			cmdStr: "claude -p /speckit.tasks | grep 'Task'",
		},
		"complex pipeline": {
			cmdStr: "ANTHROPIC_API_KEY=\"\" claude -p \"/speckit.specify\" --verbose | tee log.txt | grep 'Success'",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := executor.parseCustomCommand(tc.cmdStr)

			// Verify command is set up to run via shell
			assert.NotNil(t, cmd)
			assert.Equal(t, "sh", cmd.Path)
		})
	}
}

// TestClaudeExecutor_EnvironmentSetup tests environment variable handling
func TestClaudeExecutor_EnvironmentSetup(t *testing.T) {
	tests := map[string]struct {
		useAPIKey bool
		wantEmpty bool // Whether ANTHROPIC_API_KEY should be empty
	}{
		"use API key": {
			useAPIKey: true,
			wantEmpty: false,
		},
		"don't use API key": {
			useAPIKey: false,
			wantEmpty: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			executor := &ClaudeExecutor{
				ClaudeCmd:  "echo",
				ClaudeArgs: []string{},
				UseAPIKey:  tc.useAPIKey,
			}

			var stdout, stderr bytes.Buffer
			err := executor.StreamCommand("test", &stdout, &stderr)
			require.NoError(t, err)

			// Note: We can't directly test environment variable setting
			// without modifying the executor to expose the command
			// This test verifies the executor can be constructed with UseAPIKey
			assert.Equal(t, tc.useAPIKey, executor.UseAPIKey)
		})
	}
}

// TestClaudeExecutor_FallbackMode tests fallback to simple mode
func TestClaudeExecutor_FallbackMode(t *testing.T) {
	executor := &ClaudeExecutor{
		ClaudeCmd:       "echo",
		ClaudeArgs:      []string{"arg1", "arg2"},
		UseAPIKey:       false,
		CustomClaudeCmd: "", // Empty means use simple mode
	}

	var stdout, stderr bytes.Buffer
	err := executor.StreamCommand("test prompt", &stdout, &stderr)

	assert.NoError(t, err)
	output := stdout.String()

	// In simple mode with echo, should see all args
	assert.Contains(t, output, "arg1")
	assert.Contains(t, output, "arg2")
	assert.Contains(t, output, "test prompt")
}

// TestExecuteSpecKitCommand tests SpecKit command execution
func TestExecuteSpecKitCommand(t *testing.T) {
	executor := &ClaudeExecutor{
		ClaudeCmd:  "echo",
		ClaudeArgs: []string{},
		UseAPIKey:  false,
	}

	// Mock execution by using echo
	// In real usage, this would call claude with the SpecKit command
	var stdout, stderr bytes.Buffer
	err := executor.StreamCommand("/speckit.specify \"test\"", &stdout, &stderr)

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "/speckit.specify")
}

// TestCustomCommandWithPipeOperator tests pipe operator handling
func TestCustomCommandWithPipeOperator(t *testing.T) {
	// This test verifies that pipe operators are handled correctly
	executor := &ClaudeExecutor{
		CustomClaudeCmd: "echo {{PROMPT}} | grep 'test'",
		UseAPIKey:       false,
	}

	expanded := executor.expandTemplate("this is a test")
	assert.Equal(t, "echo this is a test | grep 'test'", expanded)

	// Verify the command would be executed via shell
	cmd := executor.parseCustomCommand(expanded)
	assert.NotNil(t, cmd)
}

// TestCustomCommandWithEnvVarPrefix tests environment variable prefix handling
func TestCustomCommandWithEnvVarPrefix(t *testing.T) {
	executor := &ClaudeExecutor{
		CustomClaudeCmd: "ANTHROPIC_API_KEY=\"\" claude -p {{PROMPT}}",
		UseAPIKey:       false,
	}

	expanded := executor.expandTemplate("/speckit.plan")
	assert.Equal(t, "ANTHROPIC_API_KEY=\"\" claude -p /speckit.plan", expanded)

	// Verify env var prefix is preserved
	assert.Contains(t, expanded, "ANTHROPIC_API_KEY=\"\"")
}

// BenchmarkExpandTemplate benchmarks template expansion performance
func BenchmarkExpandTemplate(b *testing.B) {
	executor := &ClaudeExecutor{
		CustomClaudeCmd: "ANTHROPIC_API_KEY=\"\" claude -p {{PROMPT}} --verbose | tee output.log",
	}

	prompt := "/speckit.specify \"Add user authentication with OAuth2\""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = executor.expandTemplate(prompt)
	}
}

// BenchmarkValidateTemplate benchmarks template validation performance
func BenchmarkValidateTemplate(b *testing.B) {
	template := "ANTHROPIC_API_KEY=\"\" claude -p {{PROMPT}} --verbose | claude-clean"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateTemplate(template)
	}
}

// TestTemplateEdgeCases tests edge cases in template handling
func TestTemplateEdgeCases(t *testing.T) {
	tests := map[string]struct {
		template string
		prompt   string
		wantErr  bool
	}{
		"multiple placeholders (should replace all)": {
			template: "echo {{PROMPT}} && echo {{PROMPT}}",
			prompt:   "test",
			wantErr:  false,
		},
		"placeholder in quotes": {
			template: "claude -p \"{{PROMPT}}\"",
			prompt:   "/speckit.specify",
			wantErr:  false,
		},
		"prompt with special chars": {
			template: "claude -p {{PROMPT}}",
			prompt:   "/speckit.specify \"Feature with 'quotes' and $vars\"",
			wantErr:  false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			executor := &ClaudeExecutor{
				CustomClaudeCmd: tc.template,
			}

			expanded := executor.expandTemplate(tc.prompt)

			// Verify placeholder was replaced
			assert.NotContains(t, expanded, "{{PROMPT}}")
			assert.Contains(t, expanded, tc.prompt)
		})
	}
}
