package cliagent

import (
	"context"
	"errors"
	"testing"
)

// mockConfigurableAgent is a test agent that implements Configurator.
type mockConfigurableAgent struct {
	BaseAgent
	configResult ConfigResult
	configErr    error
	callCount    int
}

func (m *mockConfigurableAgent) ConfigureProject(projectDir, specsDir string) (ConfigResult, error) {
	m.callCount++
	if m.configErr != nil {
		return ConfigResult{}, m.configErr
	}
	return m.configResult, nil
}

// mockNonConfigurableAgent is a test agent that does NOT implement Configurator.
type mockNonConfigurableAgent struct {
	BaseAgent
}

func TestConfigure(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		agent       Agent
		projectDir  string
		specsDir    string
		wantResult  *ConfigResult
		wantErr     bool
		wantErrMsg  string
		wantNilCall bool // true if Configure should return (nil, nil)
	}{
		"non-configurator agent returns nil": {
			agent:       &mockNonConfigurableAgent{},
			projectDir:  "/project",
			specsDir:    "specs",
			wantResult:  nil,
			wantNilCall: true,
		},
		"configurator with permissions added": {
			agent: &mockConfigurableAgent{
				configResult: ConfigResult{
					PermissionsAdded: []string{"Write(.autospec/**)", "Edit(.autospec/**)"},
				},
			},
			projectDir: "/project",
			specsDir:   "specs",
			wantResult: &ConfigResult{
				PermissionsAdded: []string{"Write(.autospec/**)", "Edit(.autospec/**)"},
			},
		},
		"configurator already configured": {
			agent: &mockConfigurableAgent{
				configResult: ConfigResult{
					AlreadyConfigured: true,
				},
			},
			projectDir: "/project",
			specsDir:   "specs",
			wantResult: &ConfigResult{
				AlreadyConfigured: true,
			},
		},
		"configurator with warning": {
			agent: &mockConfigurableAgent{
				configResult: ConfigResult{
					PermissionsAdded: []string{"Write(specs/**)"},
					Warning:          "Some permissions could not be added due to deny list",
				},
			},
			projectDir: "/project",
			specsDir:   "specs",
			wantResult: &ConfigResult{
				PermissionsAdded: []string{"Write(specs/**)"},
				Warning:          "Some permissions could not be added due to deny list",
			},
		},
		"configurator returns error": {
			agent: &mockConfigurableAgent{
				configErr: errors.New("permission denied"),
			},
			projectDir: "/project",
			specsDir:   "specs",
			wantErr:    true,
			wantErrMsg: "permission denied",
		},
		"custom specs directory": {
			agent: &mockConfigurableAgent{
				configResult: ConfigResult{
					PermissionsAdded: []string{"Write(features/**)", "Edit(features/**)"},
				},
			},
			projectDir: "/my/project",
			specsDir:   "features",
			wantResult: &ConfigResult{
				PermissionsAdded: []string{"Write(features/**)", "Edit(features/**)"},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := Configure(tt.agent, tt.projectDir, tt.specsDir)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.wantErrMsg != "" && err != nil {
				if err.Error() != tt.wantErrMsg {
					t.Errorf("Configure() error = %q, want %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			// Check nil call case
			if tt.wantNilCall {
				if result != nil {
					t.Errorf("Configure() result = %v, want nil for non-configurator", result)
				}
				return
			}

			// Check result
			if result == nil {
				t.Fatal("Configure() result = nil, want non-nil")
			}

			if len(result.PermissionsAdded) != len(tt.wantResult.PermissionsAdded) {
				t.Errorf("PermissionsAdded len = %d, want %d",
					len(result.PermissionsAdded), len(tt.wantResult.PermissionsAdded))
			}
			for i, perm := range result.PermissionsAdded {
				if i < len(tt.wantResult.PermissionsAdded) && perm != tt.wantResult.PermissionsAdded[i] {
					t.Errorf("PermissionsAdded[%d] = %q, want %q",
						i, perm, tt.wantResult.PermissionsAdded[i])
				}
			}
			if result.AlreadyConfigured != tt.wantResult.AlreadyConfigured {
				t.Errorf("AlreadyConfigured = %v, want %v",
					result.AlreadyConfigured, tt.wantResult.AlreadyConfigured)
			}
			if result.Warning != tt.wantResult.Warning {
				t.Errorf("Warning = %q, want %q", result.Warning, tt.wantResult.Warning)
			}
		})
	}
}

func TestIsConfigurator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		agent Agent
		want  bool
	}{
		"agent implements Configurator": {
			agent: &mockConfigurableAgent{},
			want:  true,
		},
		"agent does not implement Configurator": {
			agent: &mockNonConfigurableAgent{},
			want:  false,
		},
		"base agent does not implement Configurator": {
			agent: &BaseAgent{AgentName: "test"},
			want:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := IsConfigurator(tt.agent); got != tt.want {
				t.Errorf("IsConfigurator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigResult_ZeroValue(t *testing.T) {
	t.Parallel()

	// Verify zero value is sensible
	var result ConfigResult

	if result.AlreadyConfigured {
		t.Error("zero value AlreadyConfigured should be false")
	}
	if len(result.PermissionsAdded) != 0 {
		t.Error("zero value PermissionsAdded should be empty")
	}
	if result.Warning != "" {
		t.Error("zero value Warning should be empty string")
	}
}

func TestConfigure_IdempotencyCheck(t *testing.T) {
	t.Parallel()

	agent := &mockConfigurableAgent{
		configResult: ConfigResult{
			PermissionsAdded: []string{"Write(.autospec/**)"},
		},
	}

	// Call Configure multiple times
	for i := 0; i < 3; i++ {
		_, err := Configure(agent, "/project", "specs")
		if err != nil {
			t.Fatalf("Configure() call %d error = %v", i+1, err)
		}
	}

	// Verify ConfigureProject was called each time
	if agent.callCount != 3 {
		t.Errorf("ConfigureProject call count = %d, want 3", agent.callCount)
	}
}

// Ensure mockNonConfigurableAgent satisfies Agent interface
var _ Agent = (*mockNonConfigurableAgent)(nil)

// Ensure mockConfigurableAgent satisfies both Agent and Configurator interfaces
var _ Agent = (*mockConfigurableAgent)(nil)
var _ Configurator = (*mockConfigurableAgent)(nil)

// mockNonConfigurableAgent needs minimal Agent interface implementation
func (m *mockNonConfigurableAgent) Execute(ctx context.Context, prompt string, opts ExecOptions) (*Result, error) {
	return &Result{}, nil
}

// mockConfigurableAgent needs minimal Agent interface implementation
func (m *mockConfigurableAgent) Execute(ctx context.Context, prompt string, opts ExecOptions) (*Result, error) {
	return &Result{}, nil
}
