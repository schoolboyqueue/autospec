// Package testutil provides test utilities and helpers for autospec tests.
package testutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/ariel-frischer/autospec/internal/config"
	"github.com/ariel-frischer/autospec/internal/workflow"
)

// CallRecord records a single executor call with metadata.
type CallRecord struct {
	Method    string
	Prompt    string
	Timestamp time.Time
	Response  string
	Error     error
}

// MockExecutorBuilder provides a fluent API for configuring mock executor behavior.
type MockExecutorBuilder struct {
	responses      []mockResponse
	currentIndex   int
	calls          []CallRecord
	mu             sync.Mutex
	artifactDir    string
	mockClaudePath string
	t              *testing.T
}

type mockResponse struct {
	response    string
	responseErr error
	artifactGen func(dir string) // Function to generate artifacts
	delay       time.Duration
}

// NewMockExecutorBuilder creates a new MockExecutorBuilder for configuring mock behavior.
func NewMockExecutorBuilder(t *testing.T) *MockExecutorBuilder {
	t.Helper()

	return &MockExecutorBuilder{
		responses: make([]mockResponse, 0),
		calls:     make([]CallRecord, 0),
		t:         t,
	}
}

// WithResponse adds a successful response to the response queue.
func (b *MockExecutorBuilder) WithResponse(response string) *MockExecutorBuilder {
	b.responses = append(b.responses, mockResponse{response: response})
	return b
}

// WithError adds an error response to the response queue.
func (b *MockExecutorBuilder) WithError(err error) *MockExecutorBuilder {
	b.responses = append(b.responses, mockResponse{responseErr: err})
	return b
}

// ThenResponse adds another response to be returned on subsequent calls.
func (b *MockExecutorBuilder) ThenResponse(response string) *MockExecutorBuilder {
	return b.WithResponse(response)
}

// ThenError adds an error to be returned on subsequent calls.
func (b *MockExecutorBuilder) ThenError(err error) *MockExecutorBuilder {
	return b.WithError(err)
}

// WithDelay adds a delay before returning the response (for timeout testing).
func (b *MockExecutorBuilder) WithDelay(d time.Duration) *MockExecutorBuilder {
	if len(b.responses) > 0 {
		b.responses[len(b.responses)-1].delay = d
	}
	return b
}

// WithArtifactGeneration configures the mock to write artifacts when executed.
func (b *MockExecutorBuilder) WithArtifactGeneration(gen func(dir string)) *MockExecutorBuilder {
	if len(b.responses) > 0 {
		b.responses[len(b.responses)-1].artifactGen = gen
	}
	return b
}

// WithArtifactDir sets the directory where artifacts will be generated.
func (b *MockExecutorBuilder) WithArtifactDir(dir string) *MockExecutorBuilder {
	b.artifactDir = dir
	return b
}

// WithMockClaudePath sets the path to mock-claude.sh for integration tests.
func (b *MockExecutorBuilder) WithMockClaudePath(path string) *MockExecutorBuilder {
	b.mockClaudePath = path
	return b
}

// Build returns the configured MockExecutor.
func (b *MockExecutorBuilder) Build() *MockExecutor {
	return &MockExecutor{
		builder: b,
	}
}

// MockExecutor implements a mock executor for testing workflow operations.
type MockExecutor struct {
	builder *MockExecutorBuilder
}

// Execute simulates command execution and returns the next queued response.
func (m *MockExecutor) Execute(prompt string) error {
	return m.recordAndRespond("Execute", prompt)
}

// ExecuteSpecKitCommand simulates speckit command execution.
func (m *MockExecutor) ExecuteSpecKitCommand(command string) error {
	return m.recordAndRespond("ExecuteSpecKitCommand", command)
}

// FormatCommand returns a formatted command string.
func (m *MockExecutor) FormatCommand(prompt string) string {
	m.builder.mu.Lock()
	m.builder.calls = append(m.builder.calls, CallRecord{
		Method:    "FormatCommand",
		Prompt:    prompt,
		Timestamp: time.Now(),
	})
	m.builder.mu.Unlock()

	if m.builder.mockClaudePath != "" {
		return m.builder.mockClaudePath + " " + prompt
	}
	return "mock-claude " + prompt
}

// StreamCommand simulates streaming command execution.
func (m *MockExecutor) StreamCommand(prompt string, stdout, stderr io.Writer) error {
	err := m.recordAndRespond("StreamCommand", prompt)
	if err == nil && len(m.builder.responses) > 0 {
		// Write response to stdout
		idx := m.builder.currentIndex - 1
		if idx >= 0 && idx < len(m.builder.responses) {
			if _, writeErr := stdout.Write([]byte(m.builder.responses[idx].response)); writeErr != nil {
				return fmt.Errorf("writing to stdout: %w", writeErr)
			}
		}
	}
	return err
}

func (m *MockExecutor) recordAndRespond(method, prompt string) error {
	m.builder.mu.Lock()
	defer m.builder.mu.Unlock()

	record := CallRecord{
		Method:    method,
		Prompt:    prompt,
		Timestamp: time.Now(),
	}

	// Get next response
	if m.builder.currentIndex < len(m.builder.responses) {
		resp := m.builder.responses[m.builder.currentIndex]
		m.builder.currentIndex++

		// Apply delay if configured
		if resp.delay > 0 {
			time.Sleep(resp.delay)
		}

		// Generate artifacts if configured
		if resp.artifactGen != nil && m.builder.artifactDir != "" {
			resp.artifactGen(m.builder.artifactDir)
		}

		record.Response = resp.response
		record.Error = resp.responseErr
		m.builder.calls = append(m.builder.calls, record)

		return resp.responseErr
	}

	// No more responses configured, return success
	m.builder.calls = append(m.builder.calls, record)
	return nil
}

// GetCalls returns all recorded calls.
func (m *MockExecutor) GetCalls() []CallRecord {
	m.builder.mu.Lock()
	defer m.builder.mu.Unlock()

	result := make([]CallRecord, len(m.builder.calls))
	copy(result, m.builder.calls)
	return result
}

// GetCallCount returns the number of calls made.
func (m *MockExecutor) GetCallCount() int {
	m.builder.mu.Lock()
	defer m.builder.mu.Unlock()
	return len(m.builder.calls)
}

// GetCallsByMethod returns calls filtered by method name.
func (m *MockExecutor) GetCallsByMethod(method string) []CallRecord {
	m.builder.mu.Lock()
	defer m.builder.mu.Unlock()

	var result []CallRecord
	for _, call := range m.builder.calls {
		if call.Method == method {
			result = append(result, call)
		}
	}
	return result
}

// AssertCalled verifies that a method was called with the expected prompt.
func (m *MockExecutor) AssertCalled(t *testing.T, method, expectedPromptSubstring string) {
	t.Helper()

	calls := m.GetCallsByMethod(method)
	for _, call := range calls {
		if containsString(call.Prompt, expectedPromptSubstring) {
			return
		}
	}

	t.Errorf("expected %s to be called with prompt containing %q, but was not found in %d calls",
		method, expectedPromptSubstring, len(calls))
}

// AssertNotCalled verifies that a method was NOT called.
func (m *MockExecutor) AssertNotCalled(t *testing.T, method string) {
	t.Helper()

	calls := m.GetCallsByMethod(method)
	if len(calls) > 0 {
		t.Errorf("expected %s to not be called, but was called %d times", method, len(calls))
	}
}

// AssertCallCount verifies the number of calls to a method.
func (m *MockExecutor) AssertCallCount(t *testing.T, method string, expected int) {
	t.Helper()

	calls := m.GetCallsByMethod(method)
	if len(calls) != expected {
		t.Errorf("expected %s to be called %d times, got %d", method, expected, len(calls))
	}
}

// Reset clears all recorded calls and resets the response index.
func (m *MockExecutor) Reset() {
	m.builder.mu.Lock()
	defer m.builder.mu.Unlock()

	m.builder.calls = make([]CallRecord, 0)
	m.builder.currentIndex = 0
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ArtifactGenerators provides common artifact generation functions.
var ArtifactGenerators = struct {
	Spec  func(dir string)
	Plan  func(dir string)
	Tasks func(dir string)
}{
	Spec: func(dir string) {
		content := `feature:
  branch: "test-feature"
  created: "2025-01-01"
  status: "Draft"
  input: "test"
user_stories:
  - id: "US-001"
    title: "Test"
    priority: "P1"
    as_a: "dev"
    i_want: "test"
    so_that: "it works"
    why_this_priority: "test"
    independent_test: "test"
    acceptance_scenarios:
      - given: "a"
        when: "b"
        then: "c"
requirements:
  functional:
    - id: "FR-001"
      description: "test"
      testable: true
      acceptance_criteria: "test"
  non_functional:
    - id: "NFR-001"
      category: "code_quality"
      description: "test"
      measurable_target: "test"
success_criteria:
  measurable_outcomes:
    - id: "SC-001"
      description: "test"
      metric: "test"
      target: "test"
key_entities: []
edge_cases: []
assumptions: []
constraints: []
out_of_scope: []
_meta:
  version: "1.0.0"
  generator: "autospec"
  generator_version: "test"
  created: "2025-01-01T00:00:00Z"
  artifact_type: "spec"
`
		writeArtifact(dir, "spec.yaml", content)
	},
	Plan: func(dir string) {
		content := `plan:
  branch: "test-feature"
  created: "2025-01-01"
  spec_path: "specs/test-feature/spec.yaml"
summary: "Test plan"
technical_context:
  language: "Go"
  framework: "None"
  primary_dependencies: []
  storage: "None"
  testing:
    framework: "Go testing"
    approach: "Unit tests"
  target_platform: "Linux"
  project_type: "cli"
  performance_goals: "Fast"
  constraints: []
  scale_scope: "Small"
constitution_check:
  constitution_path: ".autospec/memory/constitution.yaml"
  gates: []
research_findings:
  decisions: []
data_model:
  entities: []
api_contracts:
  endpoints: []
project_structure:
  documentation: []
  source_code: []
  tests: []
implementation_phases:
  - phase: 1
    name: "Test"
    goal: "Test"
    deliverables: []
risks: []
open_questions: []
_meta:
  version: "1.0.0"
  generator: "autospec"
  generator_version: "test"
  created: "2025-01-01T00:00:00Z"
  artifact_type: "plan"
`
		writeArtifact(dir, "plan.yaml", content)
	},
	Tasks: func(dir string) {
		content := `tasks:
  branch: "test-feature"
  created: "2025-01-01"
  spec_path: "specs/test-feature/spec.yaml"
  plan_path: "specs/test-feature/plan.yaml"
summary:
  total_tasks: 1
  total_phases: 1
  parallel_opportunities: 0
  estimated_complexity: "low"
phases:
  - number: 1
    title: "Test"
    purpose: "Test"
    tasks:
      - id: "T001"
        title: "Test task"
        status: "Pending"
        type: "implementation"
        parallel: false
        story_id: "US-001"
        file_path: "test.go"
        dependencies: []
        acceptance_criteria:
          - "Test passes"
dependencies:
  user_story_order: []
  phase_order: []
parallel_execution: []
implementation_strategy:
  mvp_scope:
    phases: [1]
    description: "MVP"
    validation: "Tests pass"
  incremental_delivery: []
_meta:
  version: "1.0.0"
  generator: "autospec"
  generator_version: "test"
  created: "2025-01-01T00:00:00Z"
  artifact_type: "tasks"
`
		writeArtifact(dir, "tasks.yaml", content)
	},
}

func writeArtifact(dir, filename, content string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	path := filepath.Join(dir, filename)
	_ = os.WriteFile(path, []byte(content), 0644)
}
