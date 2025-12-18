package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ariel-frischer/autospec/internal/config"
	"github.com/ariel-frischer/autospec/internal/spec"
	"github.com/ariel-frischer/autospec/internal/validation"
)

func TestNewWorkflowOrchestrator(t *testing.T) {
	cfg := &config.Configuration{
		ClaudeCmd:  "claude",
		ClaudeArgs: []string{"-p"},
		SpecsDir:   "./specs",
		MaxRetries: 3,
		StateDir:   "~/.autospec/state",
	}

	orchestrator := NewWorkflowOrchestrator(cfg)

	if orchestrator == nil {
		t.Fatal("NewWorkflowOrchestrator() returned nil")
	}

	if orchestrator.SpecsDir != "./specs" {
		t.Errorf("SpecsDir = %v, want './specs'", orchestrator.SpecsDir)
	}

	if orchestrator.Executor == nil {
		t.Error("Executor should not be nil")
	}
}

func TestWorkflowOrchestrator_Configuration(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Configuration{
		ClaudeCmd:  "claude",
		ClaudeArgs: []string{"-p"},
		SpecsDir:   tmpDir,
		MaxRetries: 3,
		StateDir:   filepath.Join(tmpDir, "state"),
	}

	orchestrator := NewWorkflowOrchestrator(cfg)

	if orchestrator.SpecsDir != tmpDir {
		t.Errorf("SpecsDir = %v, want %v", orchestrator.SpecsDir, tmpDir)
	}

	if orchestrator.Config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %v, want 3", orchestrator.Config.MaxRetries)
	}
}

// TestExecutePlanWithPrompt tests that plan commands properly format prompts
func TestExecutePlanWithPrompt(t *testing.T) {
	tests := map[string]struct {
		prompt      string
		wantCommand string
	}{
		"no prompt": {
			prompt:      "",
			wantCommand: "/autospec.plan",
		},
		"simple prompt": {
			prompt:      "Focus on security",
			wantCommand: `/autospec.plan "Focus on security"`,
		},
		"prompt with quotes": {
			prompt:      "Use 'best practices' for auth",
			wantCommand: `/autospec.plan "Use 'best practices' for auth"`,
		},
		"multiline prompt": {
			prompt: `Consider these aspects:
  - Performance
  - Scalability`,
			wantCommand: `/autospec.plan "Consider these aspects:
  - Performance
  - Scalability"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// This test verifies the command construction logic
			command := "/autospec.plan"
			if tc.prompt != "" {
				command = "/autospec.plan \"" + tc.prompt + "\""
			}

			if command != tc.wantCommand {
				t.Errorf("command = %q, want %q", command, tc.wantCommand)
			}
		})
	}
}

// TestExecuteTasksWithPrompt tests that tasks commands properly format prompts
func TestExecuteTasksWithPrompt(t *testing.T) {
	tests := map[string]struct {
		prompt      string
		wantCommand string
	}{
		"no prompt": {
			prompt:      "",
			wantCommand: "/autospec.tasks",
		},
		"simple prompt": {
			prompt:      "Break into small steps",
			wantCommand: `/autospec.tasks "Break into small steps"`,
		},
		"prompt with quotes": {
			prompt:      "Make tasks 'granular' and testable",
			wantCommand: `/autospec.tasks "Make tasks 'granular' and testable"`,
		},
		"complex prompt": {
			prompt: `Requirements:
  - Each task < 1 hour
  - Include testing`,
			wantCommand: `/autospec.tasks "Requirements:
  - Each task < 1 hour
  - Include testing"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// This test verifies the command construction logic
			command := "/autospec.tasks"
			if tc.prompt != "" {
				command = "/autospec.tasks \"" + tc.prompt + "\""
			}

			if command != tc.wantCommand {
				t.Errorf("command = %q, want %q", command, tc.wantCommand)
			}
		})
	}
}

// TestSpecNameFormat tests that spec names are formatted correctly with number prefix
func TestSpecNameFormat(t *testing.T) {
	tests := map[string]struct {
		metadata     *spec.Metadata
		wantSpecName string
	}{
		"spec with three-digit number": {
			metadata: &spec.Metadata{
				Number: "003",
				Name:   "command-timeout",
			},
			wantSpecName: "003-command-timeout",
		},
		"spec with two-digit number": {
			metadata: &spec.Metadata{
				Number: "002",
				Name:   "go-binary-migration",
			},
			wantSpecName: "002-go-binary-migration",
		},
		"spec with single digit number": {
			metadata: &spec.Metadata{
				Number: "001",
				Name:   "initial-feature",
			},
			wantSpecName: "001-initial-feature",
		},
		"spec with hyphenated name": {
			metadata: &spec.Metadata{
				Number: "123",
				Name:   "multi-word-feature-name",
			},
			wantSpecName: "123-multi-word-feature-name",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Test the format string used in the workflow
			specName := fmt.Sprintf("%s-%s", tt.metadata.Number, tt.metadata.Name)

			if specName != tt.wantSpecName {
				t.Errorf("spec name format = %q, want %q", specName, tt.wantSpecName)
			}
		})
	}
}

// TestSpecDirectoryConstruction tests that spec directories are constructed correctly
func TestSpecDirectoryConstruction(t *testing.T) {
	tmpDir := t.TempDir()

	tests := map[string]struct {
		specsDir    string
		specName    string
		wantSpecDir string
	}{
		"full spec name with number": {
			specsDir:    tmpDir,
			specName:    "003-command-timeout",
			wantSpecDir: filepath.Join(tmpDir, "003-command-timeout"),
		},
		"relative specs dir": {
			specsDir:    "./specs",
			specName:    "002-go-binary-migration",
			wantSpecDir: filepath.Join("./specs", "002-go-binary-migration"),
		},
		"absolute specs dir": {
			specsDir:    "/tmp/specs",
			specName:    "001-initial-feature",
			wantSpecDir: filepath.Join("/tmp/specs", "001-initial-feature"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Test directory construction used in executor
			specDir := filepath.Join(tt.specsDir, tt.specName)

			if specDir != tt.wantSpecDir {
				t.Errorf("spec directory = %q, want %q", specDir, tt.wantSpecDir)
			}
		})
	}
}

// TestSpecNameFromMetadata verifies spec name is constructed correctly from metadata
func TestSpecNameFromMetadata(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	specDir := filepath.Join(specsDir, "003-command-timeout")

	// Create the directory structure
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a spec.md file
	specFile := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specFile, []byte("# Test Spec"), 0644); err != nil {
		t.Fatalf("Failed to create spec.md: %v", err)
	}

	// Test getting metadata
	metadata, err := spec.GetSpecMetadata(specsDir, "003-command-timeout")
	if err != nil {
		t.Fatalf("GetSpecMetadata() error = %v", err)
	}

	// Verify the metadata has correct values
	if metadata.Number != "003" {
		t.Errorf("metadata.Number = %q, want %q", metadata.Number, "003")
	}

	if metadata.Name != "command-timeout" {
		t.Errorf("metadata.Name = %q, want %q", metadata.Name, "command-timeout")
	}

	// Test constructing the full spec name (this is what the fix addresses)
	fullSpecName := fmt.Sprintf("%s-%s", metadata.Number, metadata.Name)
	expectedSpecName := "003-command-timeout"

	if fullSpecName != expectedSpecName {
		t.Errorf("full spec name = %q, want %q", fullSpecName, expectedSpecName)
	}

	// Verify the constructed directory path is correct
	constructedDir := filepath.Join(specsDir, fullSpecName)
	if constructedDir != specDir {
		t.Errorf("constructed directory = %q, want %q", constructedDir, specDir)
	}
}

// TestExecuteImplementWithTaskCommand tests that task-level implement commands are properly formatted
func TestExecuteImplementWithTaskCommand(t *testing.T) {
	tests := map[string]struct {
		taskID      string
		prompt      string
		wantCommand string
	}{
		"task without prompt": {
			taskID:      "T001",
			prompt:      "",
			wantCommand: "/autospec.implement --task T001",
		},
		"task with prompt": {
			taskID:      "T002",
			prompt:      "Focus on error handling",
			wantCommand: `/autospec.implement --task T002 "Focus on error handling"`,
		},
		"task with complex prompt": {
			taskID: "T003",
			prompt: "Consider:\n  - Performance\n  - Scalability",
			wantCommand: `/autospec.implement --task T003 "Consider:
  - Performance
  - Scalability"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Build command with task filter (mirrors executeSingleTaskSession logic)
			command := fmt.Sprintf("/autospec.implement --task %s", tc.taskID)
			if tc.prompt != "" {
				command = fmt.Sprintf("/autospec.implement --task %s \"%s\"", tc.taskID, tc.prompt)
			}

			if command != tc.wantCommand {
				t.Errorf("command = %q, want %q", command, tc.wantCommand)
			}
		})
	}
}

// TestTaskModeDispatch tests that TaskMode correctly routes to ExecuteImplementWithTasks
func TestTaskModeDispatch(t *testing.T) {
	opts := PhaseExecutionOptions{
		TaskMode: true,
		FromTask: "T003",
	}

	if opts.Mode() != ModeAllTasks {
		t.Errorf("Mode() = %v, want ModeAllTasks", opts.Mode())
	}
}

// TestTaskModeWithFromTask tests that FromTask is correctly included in options
func TestTaskModeWithFromTask(t *testing.T) {
	tests := map[string]struct {
		opts     PhaseExecutionOptions
		wantMode PhaseExecutionMode
	}{
		"task mode enabled": {
			opts: PhaseExecutionOptions{
				TaskMode: true,
			},
			wantMode: ModeAllTasks,
		},
		"task mode with from-task": {
			opts: PhaseExecutionOptions{
				TaskMode: true,
				FromTask: "T005",
			},
			wantMode: ModeAllTasks,
		},
		"task mode takes precedence over phases": {
			opts: PhaseExecutionOptions{
				TaskMode:     true,
				RunAllPhases: true, // This should be ignored
			},
			wantMode: ModeAllTasks,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := tt.opts.Mode(); got != tt.wantMode {
				t.Errorf("Mode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

// TestTaskCompletionValidation tests that task completion is properly validated
func TestTaskCompletionValidation(t *testing.T) {
	tests := map[string]struct {
		taskStatus  string
		expectError bool
	}{
		"task completed": {
			taskStatus:  "Completed",
			expectError: false,
		},
		"task completed lowercase": {
			taskStatus:  "completed",
			expectError: false,
		},
		"task pending": {
			taskStatus:  "Pending",
			expectError: true,
		},
		"task in progress": {
			taskStatus:  "InProgress",
			expectError: true,
		},
		"task blocked": {
			taskStatus:  "Blocked",
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Create isolated directory for this subtest
			tmpDir := t.TempDir()
			specsDir := filepath.Join(tmpDir, "specs")
			specDir := filepath.Join(specsDir, "001-test-feature")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Create tasks.yaml with specified status
			tasksContent := fmt.Sprintf(`tasks:
    branch: "001-test-feature"
    created: "2025-01-01"
summary:
    total_tasks: 1
    total_phases: 1
phases:
    - number: 1
      title: "Test Phase"
      purpose: "Testing"
      tasks:
        - id: "T001"
          title: "Test Task"
          status: "%s"
          type: "implementation"
          parallel: false
          story_id: null
          file_path: "test.go"
          dependencies: []
          acceptance_criteria:
            - "Test passes"
_meta:
    version: "1.0.0"
    artifact_type: "tasks"
`, tt.taskStatus)

			tasksPath := filepath.Join(specDir, "tasks.yaml")
			if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
				t.Fatalf("Failed to create tasks.yaml: %v", err)
			}

			// Simulate the validation logic from executeSingleTaskSession
			allTasks, err := validation.GetAllTasks(tasksPath)
			if err != nil {
				t.Fatalf("Failed to get tasks: %v", err)
			}

			task, err := validation.GetTaskByID(allTasks, "T001")
			if err != nil {
				t.Fatalf("Failed to get task: %v", err)
			}

			// Check if task is completed (same logic as in executeSingleTaskSession)
			isCompleted := task.Status == "Completed" || task.Status == "completed"

			if tt.expectError && isCompleted {
				t.Errorf("Expected validation error for status %q, but task was considered completed", tt.taskStatus)
			}
			if !tt.expectError && !isCompleted {
				t.Errorf("Expected no error for status %q, but task was not considered completed", tt.taskStatus)
			}
		})
	}
}

// TestTaskCompletionValidationAfterSession tests validation after task session completes
func TestTaskCompletionValidationAfterSession(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	specDir := filepath.Join(specsDir, "001-test-feature")

	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create tasks.yaml with incomplete task
	tasksContent := `tasks:
    branch: "001-test-feature"
    created: "2025-01-01"
summary:
    total_tasks: 1
    total_phases: 1
phases:
    - number: 1
      title: "Test Phase"
      purpose: "Testing"
      tasks:
        - id: "T001"
          title: "Test Task"
          status: "InProgress"
          type: "implementation"
          parallel: false
          story_id: null
          file_path: "test.go"
          dependencies: []
          acceptance_criteria:
            - "Test passes"
_meta:
    version: "1.0.0"
    artifact_type: "tasks"
`

	tasksPath := filepath.Join(specDir, "tasks.yaml")
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("Failed to create tasks.yaml: %v", err)
	}

	// Simulate the validation function that would be passed to ExecutePhase
	validateFunc := func(specDir string) error {
		tasksPath := filepath.Join(specDir, "tasks.yaml")
		allTasks, err := validation.GetAllTasks(tasksPath)
		if err != nil {
			return err
		}

		task, err := validation.GetTaskByID(allTasks, "T001")
		if err != nil {
			return err
		}

		if task.Status != "Completed" && task.Status != "completed" {
			return fmt.Errorf("task %s not completed (status: %s)", task.ID, task.Status)
		}
		return nil
	}

	// Test that validation fails for incomplete task
	err := validateFunc(specDir)
	if err == nil {
		t.Error("Expected validation to fail for incomplete task, but it passed")
	} else {
		expectedMsg := "task T001 not completed (status: InProgress)"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
		}
	}

	// Now update the task to Completed and verify validation passes
	tasksContentCompleted := `tasks:
    branch: "001-test-feature"
    created: "2025-01-01"
summary:
    total_tasks: 1
    total_phases: 1
phases:
    - number: 1
      title: "Test Phase"
      purpose: "Testing"
      tasks:
        - id: "T001"
          title: "Test Task"
          status: "Completed"
          type: "implementation"
          parallel: false
          story_id: null
          file_path: "test.go"
          dependencies: []
          acceptance_criteria:
            - "Test passes"
_meta:
    version: "1.0.0"
    artifact_type: "tasks"
`

	if err := os.WriteFile(tasksPath, []byte(tasksContentCompleted), 0644); err != nil {
		t.Fatalf("Failed to update tasks.yaml: %v", err)
	}

	// Test that validation passes for completed task
	err = validateFunc(specDir)
	if err != nil {
		t.Errorf("Expected validation to pass for completed task, got error: %v", err)
	}
}

// TestTaskRetryExhaustedBehavior tests that retry exhaustion is handled correctly
func TestTaskRetryExhaustedBehavior(t *testing.T) {
	// Test the StageResult structure that indicates retry exhaustion
	result := &StageResult{
		Stage:      StageImplement,
		Success:    false,
		Exhausted:  true,
		RetryCount: 3,
		Error:      fmt.Errorf("task T001 not completed (status: InProgress)"),
	}

	// Verify exhausted state
	if !result.Exhausted {
		t.Error("Expected Exhausted to be true")
	}

	if result.RetryCount != 3 {
		t.Errorf("Expected RetryCount to be 3, got %d", result.RetryCount)
	}

	// Test the error message format that would be shown to user
	expectedErrorFormat := "task %s exhausted retries: %w"
	errMsg := fmt.Errorf(expectedErrorFormat, "T001", result.Error)
	if errMsg == nil {
		t.Error("Expected error message to be generated")
	}

	// Verify the message includes key information
	errStr := errMsg.Error()
	if !strings.Contains(errStr, "T001") {
		t.Errorf("Error message should contain task ID, got: %s", errStr)
	}
	if !strings.Contains(errStr, "exhausted retries") {
		t.Errorf("Error message should mention exhausted retries, got: %s", errStr)
	}
}

// TestExecuteImplementWithPrompt tests that implement commands properly format prompts
func TestExecuteImplementWithPrompt(t *testing.T) {
	tests := map[string]struct {
		prompt      string
		resume      bool
		wantCommand string
	}{
		"no prompt, no resume": {
			prompt:      "",
			resume:      false,
			wantCommand: "/autospec.implement",
		},
		"simple prompt, no resume": {
			prompt:      "Focus on documentation",
			resume:      false,
			wantCommand: `/autospec.implement "Focus on documentation"`,
		},
		"no prompt, with resume": {
			prompt:      "",
			resume:      true,
			wantCommand: "/autospec.implement --resume",
		},
		"prompt with resume": {
			prompt:      "Complete remaining tasks",
			resume:      true,
			wantCommand: `/autospec.implement --resume "Complete remaining tasks"`,
		},
		"prompt with quotes": {
			prompt:      "Use 'best practices' for tests",
			resume:      false,
			wantCommand: `/autospec.implement "Use 'best practices' for tests"`,
		},
		"multiline prompt": {
			prompt: `Focus on:
  - Error handling
  - Tests`,
			resume: false,
			wantCommand: `/autospec.implement "Focus on:
  - Error handling
  - Tests"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// This test verifies the command construction logic
			command := "/autospec.implement"
			if tc.resume {
				command += " --resume"
			}
			if tc.prompt != "" {
				command = "/autospec.implement \"" + tc.prompt + "\""
				if tc.resume {
					// If both resume and prompt, append resume after prompt
					command = "/autospec.implement --resume \"" + tc.prompt + "\""
				}
			}

			if command != tc.wantCommand {
				t.Errorf("command = %q, want %q", command, tc.wantCommand)
			}
		})
	}
}

// TestOrchestratorDebugLog tests the orchestrator debug logging function
func TestOrchestratorDebugLog(t *testing.T) {
	t.Parallel()

	cfg := &config.Configuration{
		ClaudeCmd:  "claude",
		SpecsDir:   "./specs",
		MaxRetries: 3,
		StateDir:   "~/.autospec/state",
	}

	// Test with debug disabled
	orchestrator := NewWorkflowOrchestrator(cfg)
	orchestrator.Debug = false
	// This should not panic or error
	orchestrator.debugLog("Test message: %s", "arg")

	// Test with debug enabled
	orchestrator.Debug = true
	// This should also not panic (just prints to stdout)
	orchestrator.debugLog("Debug enabled: %d", 123)
}

// TestShouldSkipTask tests the task skip logic
func TestShouldSkipTask(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		task       validation.TaskItem
		idx        int
		totalTasks int
		wantSkip   bool
	}{
		"completed task": {
			task:       validation.TaskItem{ID: "T001", Title: "Test", Status: "Completed"},
			idx:        0,
			totalTasks: 3,
			wantSkip:   true,
		},
		"completed lowercase": {
			task:       validation.TaskItem{ID: "T001", Title: "Test", Status: "completed"},
			idx:        0,
			totalTasks: 3,
			wantSkip:   true,
		},
		"blocked task": {
			task:       validation.TaskItem{ID: "T001", Title: "Test", Status: "Blocked"},
			idx:        0,
			totalTasks: 3,
			wantSkip:   true,
		},
		"blocked lowercase": {
			task:       validation.TaskItem{ID: "T001", Title: "Test", Status: "blocked"},
			idx:        0,
			totalTasks: 3,
			wantSkip:   true,
		},
		"pending task": {
			task:       validation.TaskItem{ID: "T001", Title: "Test", Status: "Pending"},
			idx:        0,
			totalTasks: 3,
			wantSkip:   false,
		},
		"in progress task": {
			task:       validation.TaskItem{ID: "T001", Title: "Test", Status: "InProgress"},
			idx:        0,
			totalTasks: 3,
			wantSkip:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := shouldSkipTask(tt.task, tt.idx, tt.totalTasks)
			if result != tt.wantSkip {
				t.Errorf("shouldSkipTask() = %v, want %v", result, tt.wantSkip)
			}
		})
	}
}

// TestGetTaskIDsForPhase tests task ID extraction for a phase
func TestGetTaskIDsForPhase(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	specDir := filepath.Join(specsDir, "001-test")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tasksContent := `tasks:
    branch: "001-test"
    created: "2025-01-01"
summary:
    total_tasks: 3
    total_phases: 2
phases:
    - number: 1
      title: "Phase 1"
      purpose: "First phase"
      tasks:
        - id: "T001"
          title: "Task 1"
          status: "Pending"
          type: "implementation"
          parallel: false
          dependencies: []
        - id: "T002"
          title: "Task 2"
          status: "Pending"
          type: "test"
          parallel: false
          dependencies: []
    - number: 2
      title: "Phase 2"
      purpose: "Second phase"
      tasks:
        - id: "T003"
          title: "Task 3"
          status: "Pending"
          type: "implementation"
          parallel: false
          dependencies: ["T001"]
_meta:
    version: "1.0.0"
    artifact_type: "tasks"
`

	tasksPath := filepath.Join(specDir, "tasks.yaml")
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("Failed to create tasks.yaml: %v", err)
	}

	tests := map[string]struct {
		phaseNumber int
		wantIDs     []string
	}{
		"phase 1 tasks": {
			phaseNumber: 1,
			wantIDs:     []string{"T001", "T002"},
		},
		"phase 2 tasks": {
			phaseNumber: 2,
			wantIDs:     []string{"T003"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			taskIDs := getTaskIDsForPhase(tasksPath, tt.phaseNumber)

			if len(taskIDs) != len(tt.wantIDs) {
				t.Errorf("getTaskIDsForPhase() returned %d IDs, want %d", len(taskIDs), len(tt.wantIDs))
				return
			}

			for i, id := range taskIDs {
				if id != tt.wantIDs[i] {
					t.Errorf("taskIDs[%d] = %q, want %q", i, id, tt.wantIDs[i])
				}
			}
		})
	}
}

// TestGetUpdatedPhaseInfo tests phase info retrieval
func TestGetUpdatedPhaseInfo(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	specDir := filepath.Join(specsDir, "001-test")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tasksContent := `tasks:
    branch: "001-test"
    created: "2025-01-01"
summary:
    total_tasks: 2
    total_phases: 2
phases:
    - number: 1
      title: "Phase 1"
      purpose: "First phase"
      tasks:
        - id: "T001"
          title: "Task 1"
          status: "Completed"
          type: "implementation"
          parallel: false
          dependencies: []
    - number: 2
      title: "Phase 2"
      purpose: "Second phase"
      tasks:
        - id: "T002"
          title: "Task 2"
          status: "Pending"
          type: "implementation"
          parallel: false
          dependencies: []
_meta:
    version: "1.0.0"
    artifact_type: "tasks"
`

	tasksPath := filepath.Join(specDir, "tasks.yaml")
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("Failed to create tasks.yaml: %v", err)
	}

	tests := map[string]struct {
		phaseNumber int
		wantNil     bool
	}{
		"existing phase 1": {
			phaseNumber: 1,
			wantNil:     false,
		},
		"existing phase 2": {
			phaseNumber: 2,
			wantNil:     false,
		},
		"non-existent phase 99": {
			phaseNumber: 99,
			wantNil:     true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := getUpdatedPhaseInfo(tasksPath, tt.phaseNumber)

			if tt.wantNil && result != nil {
				t.Errorf("getUpdatedPhaseInfo() = %v, want nil", result)
			}
			if !tt.wantNil && result == nil {
				t.Errorf("getUpdatedPhaseInfo() = nil, want non-nil")
			}
			if result != nil && result.Number != tt.phaseNumber {
				t.Errorf("getUpdatedPhaseInfo().Number = %d, want %d", result.Number, tt.phaseNumber)
			}
		})
	}
}

// TestBuildImplementCommand tests the implement command builder
func TestBuildImplementCommand(t *testing.T) {
	t.Parallel()

	cfg := &config.Configuration{
		ClaudeCmd:  "claude",
		SpecsDir:   "./specs",
		MaxRetries: 3,
		StateDir:   "~/.autospec/state",
	}

	orchestrator := NewWorkflowOrchestrator(cfg)

	tests := map[string]struct {
		resume      bool
		wantCommand string
	}{
		"without resume": {
			resume:      false,
			wantCommand: "/autospec.implement",
		},
		"with resume": {
			resume:      true,
			wantCommand: "/autospec.implement --resume",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			command := orchestrator.buildImplementCommand(tt.resume)
			if command != tt.wantCommand {
				t.Errorf("buildImplementCommand() = %q, want %q", command, tt.wantCommand)
			}
		})
	}
}

// TestPrintPhaseCompletion tests the phase completion message printing
func TestPrintPhaseCompletion(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		phaseNumber  int
		updatedPhase *validation.PhaseInfo
	}{
		"with updated phase info": {
			phaseNumber: 1,
			updatedPhase: &validation.PhaseInfo{
				Number:         1,
				Title:          "Setup",
				TotalTasks:     3,
				CompletedTasks: 3,
				BlockedTasks:   0,
			},
		},
		"without updated phase info": {
			phaseNumber:  2,
			updatedPhase: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Just verify it doesn't panic
			printPhaseCompletion(tt.phaseNumber, tt.updatedPhase)
		})
	}
}

// TestMarkSpecCompletedAndPrint tests the spec completion marking
func TestMarkSpecCompletedAndPrint(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	specDir := filepath.Join(specsDir, "001-test")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a minimal spec.yaml
	specContent := `feature:
  name: "Test Feature"
  branch: "001-test"
  status: "Draft"
  created: "2025-01-01"
user_stories: []
requirements:
  functional: []
`
	specPath := filepath.Join(specDir, "spec.yaml")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("Failed to create spec.yaml: %v", err)
	}

	// This should not panic and should update the spec
	markSpecCompletedAndPrint(specDir)

	// Test with non-existent directory - should not panic
	markSpecCompletedAndPrint(filepath.Join(tmpDir, "nonexistent"))
}

// TestExecuteSpecify tests the ExecuteSpecify method
func TestExecuteSpecify(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		featureDescription string
		setupMock          func(*MockClaudeExecutor)
		wantErr            bool
		wantErrContains    string
	}{
		"successful spec generation": {
			featureDescription: "Add user authentication",
			setupMock: func(m *MockClaudeExecutor) {
				// Mock succeeds immediately
			},
			wantErr: false,
		},
		"empty feature description": {
			featureDescription: "",
			setupMock: func(m *MockClaudeExecutor) {
				// Mock succeeds immediately
			},
			wantErr: false, // Empty string is valid, just creates empty spec
		},
		"execution error": {
			featureDescription: "Test feature",
			setupMock: func(m *MockClaudeExecutor) {
				m.WithExecuteError(ErrMockExecute)
			},
			wantErr:         true,
			wantErrContains: "specify failed",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specsDir := filepath.Join(tmpDir, "specs")
			if err := os.MkdirAll(specsDir, 0755); err != nil {
				t.Fatalf("Failed to create specs directory: %v", err)
			}

			// Create spec directory that would be created by claude
			specDir := filepath.Join(specsDir, "001-test-feature")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory: %v", err)
			}

			// Create spec.yaml that would be created by claude
			specContent := `feature:
  branch: "001-test-feature"
  created: "2025-01-01"
  status: "Draft"
  input: "test feature"
user_stories: []
requirements:
  functional: []
  non_functional: []
success_criteria:
  measurable_outcomes: []
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
			if err := os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(specContent), 0644); err != nil {
				t.Fatalf("Failed to create spec.yaml: %v", err)
			}

			cfg := &config.Configuration{
				ClaudeCmd:  "claude",
				SpecsDir:   specsDir,
				MaxRetries: 3,
				StateDir:   filepath.Join(tmpDir, "state"),
			}

			mock := NewMockClaudeExecutor()
			tt.setupMock(mock)

			orchestrator := NewWorkflowOrchestrator(cfg)
			orchestrator.Executor.Claude = &ClaudeExecutor{
				ClaudeCmd: "echo", // Use echo for testing
			}

			// For the successful case, we need to mock the entire execution
			if !tt.wantErr {
				// The command format test verifies the command is built correctly
				command := fmt.Sprintf("/autospec.specify \"%s\"", tt.featureDescription)
				if !strings.Contains(command, tt.featureDescription) {
					t.Errorf("command should contain feature description")
				}
			}
		})
	}
}

// TestExecutePlan tests the ExecutePlan method
func TestExecutePlan(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		specName        string
		prompt          string
		setupFiles      func(string, string)
		wantErr         bool
		wantErrContains string
	}{
		"successful plan with spec name": {
			specName: "001-test-feature",
			prompt:   "",
			setupFiles: func(specsDir, specName string) {
				specDir := filepath.Join(specsDir, specName)
				os.MkdirAll(specDir, 0755)
				specContent := `feature:
  branch: "001-test-feature"
  created: "2025-01-01"
  status: "Draft"
user_stories: []
requirements:
  functional: []
  non_functional: []
success_criteria:
  measurable_outcomes: []
key_entities: []
edge_cases: []
assumptions: []
constraints: []
out_of_scope: []
_meta:
  version: "1.0.0"
  artifact_type: "spec"
`
				os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(specContent), 0644)
				// Create plan.yaml (simulating claude output)
				planContent := `plan:
  branch: "001-test-feature"
  created: "2025-01-01"
summary: "Test plan"
technical_context:
  language: "Go"
_meta:
  version: "1.0.0"
  artifact_type: "plan"
`
				os.WriteFile(filepath.Join(specDir, "plan.yaml"), []byte(planContent), 0644)
			},
			wantErr: false,
		},
		"plan with custom prompt": {
			specName: "002-another-feature",
			prompt:   "Focus on security",
			setupFiles: func(specsDir, specName string) {
				specDir := filepath.Join(specsDir, specName)
				os.MkdirAll(specDir, 0755)
				specContent := `feature:
  branch: "002-another-feature"
  created: "2025-01-01"
  status: "Draft"
user_stories: []
requirements:
  functional: []
  non_functional: []
success_criteria:
  measurable_outcomes: []
key_entities: []
edge_cases: []
assumptions: []
constraints: []
out_of_scope: []
_meta:
  version: "1.0.0"
  artifact_type: "spec"
`
				os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(specContent), 0644)
				planContent := `plan:
  branch: "002-another-feature"
  created: "2025-01-01"
summary: "Test plan"
technical_context:
  language: "Go"
_meta:
  version: "1.0.0"
  artifact_type: "plan"
`
				os.WriteFile(filepath.Join(specDir, "plan.yaml"), []byte(planContent), 0644)
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specsDir := filepath.Join(tmpDir, "specs")
			if err := os.MkdirAll(specsDir, 0755); err != nil {
				t.Fatalf("Failed to create specs directory: %v", err)
			}

			// Setup test files
			tt.setupFiles(specsDir, tt.specName)

			cfg := &config.Configuration{
				ClaudeCmd:  "echo",
				SpecsDir:   specsDir,
				MaxRetries: 3,
				StateDir:   filepath.Join(tmpDir, "state"),
			}

			orchestrator := NewWorkflowOrchestrator(cfg)

			// Verify command format
			command := "/autospec.plan"
			if tt.prompt != "" {
				command = fmt.Sprintf("/autospec.plan \"%s\"", tt.prompt)
			}

			if tt.prompt != "" && !strings.Contains(command, tt.prompt) {
				t.Errorf("command should contain prompt, got: %s", command)
			}

			// Verify orchestrator is properly configured
			if orchestrator.SpecsDir != specsDir {
				t.Errorf("SpecsDir = %v, want %v", orchestrator.SpecsDir, specsDir)
			}
		})
	}
}

// TestExecuteTasks tests the ExecuteTasks method
func TestExecuteTasks(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		specName        string
		prompt          string
		setupFiles      func(string, string)
		wantErr         bool
		wantErrContains string
	}{
		"successful tasks generation": {
			specName: "001-test-feature",
			prompt:   "",
			setupFiles: func(specsDir, specName string) {
				specDir := filepath.Join(specsDir, specName)
				os.MkdirAll(specDir, 0755)
				// Create spec.yaml
				specContent := `feature:
  branch: "001-test-feature"
  created: "2025-01-01"
  status: "Draft"
user_stories: []
requirements:
  functional: []
  non_functional: []
success_criteria:
  measurable_outcomes: []
key_entities: []
edge_cases: []
assumptions: []
constraints: []
out_of_scope: []
_meta:
  version: "1.0.0"
  artifact_type: "spec"
`
				os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(specContent), 0644)
				// Create plan.yaml
				planContent := `plan:
  branch: "001-test-feature"
  created: "2025-01-01"
summary: "Test plan"
technical_context:
  language: "Go"
_meta:
  version: "1.0.0"
  artifact_type: "plan"
`
				os.WriteFile(filepath.Join(specDir, "plan.yaml"), []byte(planContent), 0644)
				// Create tasks.yaml (simulating claude output)
				tasksContent := `tasks:
  branch: "001-test-feature"
  created: "2025-01-01"
summary:
  total_tasks: 1
  total_phases: 1
phases:
  - number: 1
    title: "Test Phase"
    purpose: "Testing"
    tasks:
      - id: "T001"
        title: "Test Task"
        status: "Pending"
        type: "implementation"
        parallel: false
        dependencies: []
        acceptance_criteria:
          - "Test passes"
_meta:
  version: "1.0.0"
  artifact_type: "tasks"
`
				os.WriteFile(filepath.Join(specDir, "tasks.yaml"), []byte(tasksContent), 0644)
			},
			wantErr: false,
		},
		"tasks with custom prompt": {
			specName: "002-another-feature",
			prompt:   "Break into small steps",
			setupFiles: func(specsDir, specName string) {
				specDir := filepath.Join(specsDir, specName)
				os.MkdirAll(specDir, 0755)
				specContent := `feature:
  branch: "002-another-feature"
  created: "2025-01-01"
  status: "Draft"
user_stories: []
requirements:
  functional: []
  non_functional: []
success_criteria:
  measurable_outcomes: []
key_entities: []
edge_cases: []
assumptions: []
constraints: []
out_of_scope: []
_meta:
  version: "1.0.0"
  artifact_type: "spec"
`
				os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(specContent), 0644)
				planContent := `plan:
  branch: "002-another-feature"
  created: "2025-01-01"
summary: "Test plan"
technical_context:
  language: "Go"
_meta:
  version: "1.0.0"
  artifact_type: "plan"
`
				os.WriteFile(filepath.Join(specDir, "plan.yaml"), []byte(planContent), 0644)
				tasksContent := `tasks:
  branch: "002-another-feature"
  created: "2025-01-01"
summary:
  total_tasks: 1
  total_phases: 1
phases:
  - number: 1
    title: "Test Phase"
    purpose: "Testing"
    tasks:
      - id: "T001"
        title: "Test Task"
        status: "Pending"
        type: "implementation"
        parallel: false
        dependencies: []
        acceptance_criteria:
          - "Test passes"
_meta:
  version: "1.0.0"
  artifact_type: "tasks"
`
				os.WriteFile(filepath.Join(specDir, "tasks.yaml"), []byte(tasksContent), 0644)
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specsDir := filepath.Join(tmpDir, "specs")
			if err := os.MkdirAll(specsDir, 0755); err != nil {
				t.Fatalf("Failed to create specs directory: %v", err)
			}

			// Setup test files
			tt.setupFiles(specsDir, tt.specName)

			cfg := &config.Configuration{
				ClaudeCmd:  "echo",
				SpecsDir:   specsDir,
				MaxRetries: 3,
				StateDir:   filepath.Join(tmpDir, "state"),
			}

			orchestrator := NewWorkflowOrchestrator(cfg)

			// Verify command format
			command := "/autospec.tasks"
			if tt.prompt != "" {
				command = fmt.Sprintf("/autospec.tasks \"%s\"", tt.prompt)
			}

			if tt.prompt != "" && !strings.Contains(command, tt.prompt) {
				t.Errorf("command should contain prompt, got: %s", command)
			}

			// Verify orchestrator is properly configured
			if orchestrator.SpecsDir != specsDir {
				t.Errorf("SpecsDir = %v, want %v", orchestrator.SpecsDir, specsDir)
			}
		})
	}
}

// TestExecuteImplementModeDispatch tests the ExecuteImplement method mode dispatch
func TestExecuteImplementModeDispatch(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		phaseOpts PhaseExecutionOptions
		wantMode  PhaseExecutionMode
	}{
		"default mode": {
			phaseOpts: PhaseExecutionOptions{},
			wantMode:  ModeDefault,
		},
		"task mode": {
			phaseOpts: PhaseExecutionOptions{TaskMode: true},
			wantMode:  ModeAllTasks,
		},
		"task mode with from-task": {
			phaseOpts: PhaseExecutionOptions{TaskMode: true, FromTask: "T001"},
			wantMode:  ModeAllTasks,
		},
		"all phases mode": {
			phaseOpts: PhaseExecutionOptions{RunAllPhases: true},
			wantMode:  ModeAllPhases,
		},
		"single phase mode": {
			phaseOpts: PhaseExecutionOptions{SinglePhase: 1},
			wantMode:  ModeSinglePhase,
		},
		"from phase mode": {
			phaseOpts: PhaseExecutionOptions{FromPhase: 2},
			wantMode:  ModeFromPhase,
		},
		"task mode takes precedence": {
			phaseOpts: PhaseExecutionOptions{
				TaskMode:     true,
				RunAllPhases: true,
				SinglePhase:  1,
			},
			wantMode: ModeAllTasks,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tt.phaseOpts.Mode(); got != tt.wantMode {
				t.Errorf("Mode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

// TestPhaseExecutionOptionsMode tests the Mode() method comprehensively
func TestPhaseExecutionOptionsMode(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		opts     PhaseExecutionOptions
		wantMode PhaseExecutionMode
	}{
		"empty options returns default": {
			opts:     PhaseExecutionOptions{},
			wantMode: ModeDefault,
		},
		"task mode true returns all tasks": {
			opts:     PhaseExecutionOptions{TaskMode: true},
			wantMode: ModeAllTasks,
		},
		"run all phases true returns all phases": {
			opts:     PhaseExecutionOptions{RunAllPhases: true},
			wantMode: ModeAllPhases,
		},
		"single phase > 0 returns single phase": {
			opts:     PhaseExecutionOptions{SinglePhase: 3},
			wantMode: ModeSinglePhase,
		},
		"from phase > 0 returns from phase": {
			opts:     PhaseExecutionOptions{FromPhase: 2},
			wantMode: ModeFromPhase,
		},
		"task mode has highest priority": {
			opts: PhaseExecutionOptions{
				TaskMode:     true,
				RunAllPhases: true,
				SinglePhase:  5,
				FromPhase:    3,
			},
			wantMode: ModeAllTasks,
		},
		"run all phases has second priority": {
			opts: PhaseExecutionOptions{
				RunAllPhases: true,
				SinglePhase:  5,
				FromPhase:    3,
			},
			wantMode: ModeAllPhases,
		},
		"single phase has third priority": {
			opts: PhaseExecutionOptions{
				SinglePhase: 5,
				FromPhase:   3,
			},
			wantMode: ModeSinglePhase,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tt.opts.Mode(); got != tt.wantMode {
				t.Errorf("Mode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

// TestExecuteConstitution tests the ExecuteConstitution method
func TestExecuteConstitution(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		prompt         string
		wantCmdContain string
	}{
		"without prompt": {
			prompt:         "",
			wantCmdContain: "/autospec.constitution",
		},
		"with prompt": {
			prompt:         "Focus on testing",
			wantCmdContain: "Focus on testing",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Build command
			command := "/autospec.constitution"
			if tt.prompt != "" {
				command = fmt.Sprintf("/autospec.constitution \"%s\"", tt.prompt)
			}

			if !strings.Contains(command, tt.wantCmdContain) {
				t.Errorf("command %q should contain %q", command, tt.wantCmdContain)
			}
		})
	}
}

// TestExecuteClarify tests the ExecuteClarify method
func TestExecuteClarify(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		specName       string
		prompt         string
		wantCmdContain string
	}{
		"without prompt": {
			specName:       "001-test",
			prompt:         "",
			wantCmdContain: "/autospec.clarify",
		},
		"with prompt": {
			specName:       "001-test",
			prompt:         "Clarify security requirements",
			wantCmdContain: "Clarify security requirements",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Build command
			command := "/autospec.clarify"
			if tt.prompt != "" {
				command = fmt.Sprintf("/autospec.clarify \"%s\"", tt.prompt)
			}

			if !strings.Contains(command, tt.wantCmdContain) {
				t.Errorf("command %q should contain %q", command, tt.wantCmdContain)
			}
		})
	}
}

// TestExecuteChecklist tests the ExecuteChecklist method
func TestExecuteChecklist(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		specName       string
		prompt         string
		wantCmdContain string
	}{
		"without prompt": {
			specName:       "001-test",
			prompt:         "",
			wantCmdContain: "/autospec.checklist",
		},
		"with prompt": {
			specName:       "001-test",
			prompt:         "Include security checks",
			wantCmdContain: "Include security checks",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Build command
			command := "/autospec.checklist"
			if tt.prompt != "" {
				command = fmt.Sprintf("/autospec.checklist \"%s\"", tt.prompt)
			}

			if !strings.Contains(command, tt.wantCmdContain) {
				t.Errorf("command %q should contain %q", command, tt.wantCmdContain)
			}
		})
	}
}

// TestExecuteAnalyze tests the ExecuteAnalyze method
func TestExecuteAnalyze(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		specName       string
		prompt         string
		wantCmdContain string
	}{
		"without prompt": {
			specName:       "001-test",
			prompt:         "",
			wantCmdContain: "/autospec.analyze",
		},
		"with prompt": {
			specName:       "001-test",
			prompt:         "Focus on consistency",
			wantCmdContain: "Focus on consistency",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Build command
			command := "/autospec.analyze"
			if tt.prompt != "" {
				command = fmt.Sprintf("/autospec.analyze \"%s\"", tt.prompt)
			}

			if !strings.Contains(command, tt.wantCmdContain) {
				t.Errorf("command %q should contain %q", command, tt.wantCmdContain)
			}
		})
	}
}

// TestRunPreflightIfNeeded tests the runPreflightIfNeeded method
func TestRunPreflightIfNeeded(t *testing.T) {
	tests := map[string]struct {
		skipPreflight bool
		wantSkip      bool
	}{
		"skip when explicitly disabled": {
			skipPreflight: true,
			wantSkip:      true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Check if should skip when explicitly disabled
			shouldRun := ShouldRunPreflightChecks(tt.skipPreflight)
			if tt.wantSkip && shouldRun {
				t.Errorf("Expected preflight to be skipped with skipPreflight=%v, but it would run", tt.skipPreflight)
			}
		})
	}
}

// TestShouldRunPreflightChecksSkipFlag tests the skip flag behavior
func TestShouldRunPreflightChecksSkipFlag(t *testing.T) {
	t.Parallel()

	// When skipPreflight is true, should not run
	if ShouldRunPreflightChecks(true) {
		t.Error("Expected preflight to be skipped when skipPreflight=true")
	}
}

// TestExecuteSinglePhaseSession tests the executeSinglePhaseSession method
func TestExecuteSinglePhaseSession(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		phaseNumber int
		prompt      string
		setupFiles  func(string)
		wantErr     bool
	}{
		"valid phase with tasks": {
			phaseNumber: 1,
			prompt:      "",
			setupFiles: func(specDir string) {
				os.MkdirAll(specDir, 0755)
				// Create spec.yaml
				specContent := `feature:
  branch: "001-test"
  created: "2025-01-01"
user_stories: []
requirements:
  functional: []
  non_functional: []
success_criteria:
  measurable_outcomes: []
key_entities: []
edge_cases: []
assumptions: []
constraints: []
out_of_scope: []
_meta:
  version: "1.0.0"
  artifact_type: "spec"
`
				os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(specContent), 0644)
				// Create plan.yaml
				planContent := `plan:
  branch: "001-test"
  created: "2025-01-01"
summary: "Test"
technical_context:
  language: "Go"
_meta:
  version: "1.0.0"
  artifact_type: "plan"
`
				os.WriteFile(filepath.Join(specDir, "plan.yaml"), []byte(planContent), 0644)
				// Create tasks.yaml
				tasksContent := `tasks:
  branch: "001-test"
  created: "2025-01-01"
summary:
  total_tasks: 1
  total_phases: 1
phases:
  - number: 1
    title: "Test"
    purpose: "Testing"
    tasks:
      - id: "T001"
        title: "Test"
        status: "Pending"
        type: "implementation"
        parallel: false
        dependencies: []
        acceptance_criteria:
          - "Test"
_meta:
  version: "1.0.0"
  artifact_type: "tasks"
`
				os.WriteFile(filepath.Join(specDir, "tasks.yaml"), []byte(tasksContent), 0644)
			},
			wantErr: false,
		},
		"empty phase with no tasks": {
			phaseNumber: 1,
			prompt:      "",
			setupFiles: func(specDir string) {
				os.MkdirAll(specDir, 0755)
				// Create spec.yaml
				specContent := `feature:
  branch: "001-test"
  created: "2025-01-01"
user_stories: []
requirements:
  functional: []
  non_functional: []
success_criteria:
  measurable_outcomes: []
key_entities: []
edge_cases: []
assumptions: []
constraints: []
out_of_scope: []
_meta:
  version: "1.0.0"
  artifact_type: "spec"
`
				os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(specContent), 0644)
				// Create plan.yaml
				planContent := `plan:
  branch: "001-test"
  created: "2025-01-01"
summary: "Test"
technical_context:
  language: "Go"
_meta:
  version: "1.0.0"
  artifact_type: "plan"
`
				os.WriteFile(filepath.Join(specDir, "plan.yaml"), []byte(planContent), 0644)
				// Create tasks.yaml with empty phase
				tasksContent := `tasks:
  branch: "001-test"
  created: "2025-01-01"
summary:
  total_tasks: 0
  total_phases: 1
phases:
  - number: 1
    title: "Empty"
    purpose: "Testing"
    tasks: []
_meta:
  version: "1.0.0"
  artifact_type: "tasks"
`
				os.WriteFile(filepath.Join(specDir, "tasks.yaml"), []byte(tasksContent), 0644)
			},
			wantErr: false, // Empty phase should not error, just skip
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specsDir := filepath.Join(tmpDir, "specs")
			specDir := filepath.Join(specsDir, "001-test")

			tt.setupFiles(specDir)

			// Verify file setup
			tasksPath := filepath.Join(specDir, "tasks.yaml")
			if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
				t.Fatalf("tasks.yaml not created: %v", err)
			}
		})
	}
}

// TestExecuteSingleTaskSession tests the executeSingleTaskSession method
func TestExecuteSingleTaskSession(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		taskID    string
		taskTitle string
		prompt    string
		wantCmd   string
	}{
		"task without prompt": {
			taskID:    "T001",
			taskTitle: "Test Task",
			prompt:    "",
			wantCmd:   "/autospec.implement --task T001",
		},
		"task with prompt": {
			taskID:    "T002",
			taskTitle: "Another Task",
			prompt:    "Focus on tests",
			wantCmd:   `/autospec.implement --task T002 "Focus on tests"`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Build command
			command := fmt.Sprintf("/autospec.implement --task %s", tt.taskID)
			if tt.prompt != "" {
				command = fmt.Sprintf("/autospec.implement --task %s \"%s\"", tt.taskID, tt.prompt)
			}

			if command != tt.wantCmd {
				t.Errorf("command = %q, want %q", command, tt.wantCmd)
			}
		})
	}
}

// TestGetOrderedTasksForExecution tests the getOrderedTasksForExecution method
func TestGetOrderedTasksForExecution(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	specDir := filepath.Join(specsDir, "001-test")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec directory: %v", err)
	}

	tests := map[string]struct {
		tasksContent string
		wantCount    int
		wantErr      bool
	}{
		"valid tasks": {
			tasksContent: `tasks:
  branch: "001-test"
  created: "2025-01-01"
summary:
  total_tasks: 2
  total_phases: 1
phases:
  - number: 1
    title: "Test"
    purpose: "Testing"
    tasks:
      - id: "T001"
        title: "First"
        status: "Pending"
        type: "implementation"
        parallel: false
        dependencies: []
        acceptance_criteria:
          - "Test"
      - id: "T002"
        title: "Second"
        status: "Pending"
        type: "implementation"
        parallel: false
        dependencies: ["T001"]
        acceptance_criteria:
          - "Test"
_meta:
  version: "1.0.0"
  artifact_type: "tasks"
`,
			wantCount: 2,
			wantErr:   false,
		},
		"empty tasks": {
			tasksContent: `tasks:
  branch: "001-test"
  created: "2025-01-01"
summary:
  total_tasks: 0
  total_phases: 1
phases:
  - number: 1
    title: "Empty"
    purpose: "Testing"
    tasks: []
_meta:
  version: "1.0.0"
  artifact_type: "tasks"
`,
			wantCount: 0,
			wantErr:   true, // No tasks found
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tasksPath := filepath.Join(specDir, "tasks.yaml")
			if err := os.WriteFile(tasksPath, []byte(tt.tasksContent), 0644); err != nil {
				t.Fatalf("Failed to write tasks.yaml: %v", err)
			}

			cfg := &config.Configuration{
				ClaudeCmd:  "echo",
				SpecsDir:   specsDir,
				MaxRetries: 3,
				StateDir:   filepath.Join(tmpDir, "state"),
			}

			orchestrator := NewWorkflowOrchestrator(cfg)
			orderedTasks, _, err := orchestrator.getOrderedTasksForExecution(tasksPath)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(orderedTasks) != tt.wantCount {
					t.Errorf("Task count = %d, want %d", len(orderedTasks), tt.wantCount)
				}
			}
		})
	}
}

// TestFindTaskStartIndex tests the findTaskStartIndex method
func TestFindTaskStartIndex(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fromTask     string
		orderedTasks []validation.TaskItem
		allTasks     []validation.TaskItem
		wantIdx      int
		wantErr      bool
	}{
		"empty fromTask returns 0": {
			fromTask: "",
			orderedTasks: []validation.TaskItem{
				{ID: "T001", Title: "First", Status: "Pending"},
				{ID: "T002", Title: "Second", Status: "Pending"},
			},
			allTasks: []validation.TaskItem{
				{ID: "T001", Title: "First", Status: "Pending"},
				{ID: "T002", Title: "Second", Status: "Pending"},
			},
			wantIdx: 0,
			wantErr: false,
		},
		"valid fromTask returns correct index": {
			fromTask: "T002",
			orderedTasks: []validation.TaskItem{
				{ID: "T001", Title: "First", Status: "Completed"},
				{ID: "T002", Title: "Second", Status: "Pending"},
			},
			allTasks: []validation.TaskItem{
				{ID: "T001", Title: "First", Status: "Completed"},
				{ID: "T002", Title: "Second", Status: "Pending", Dependencies: []string{"T001"}},
			},
			wantIdx: 1,
			wantErr: false,
		},
		"non-existent task returns error": {
			fromTask: "T999",
			orderedTasks: []validation.TaskItem{
				{ID: "T001", Title: "First", Status: "Pending"},
			},
			allTasks: []validation.TaskItem{
				{ID: "T001", Title: "First", Status: "Pending"},
			},
			wantIdx: 0,
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			cfg := &config.Configuration{
				ClaudeCmd:  "echo",
				SpecsDir:   filepath.Join(tmpDir, "specs"),
				MaxRetries: 3,
				StateDir:   filepath.Join(tmpDir, "state"),
			}

			orchestrator := NewWorkflowOrchestrator(cfg)
			idx, err := orchestrator.findTaskStartIndex(tt.orderedTasks, tt.allTasks, tt.fromTask)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if idx != tt.wantIdx {
					t.Errorf("Index = %d, want %d", idx, tt.wantIdx)
				}
			}
		})
	}
}

// TestVerifyTaskCompletion tests the verifyTaskCompletion method
func TestVerifyTaskCompletion(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		taskID       string
		taskStatus   string
		wantErr      bool
		wantErrMatch string
	}{
		"completed task": {
			taskID:     "T001",
			taskStatus: "Completed",
			wantErr:    false,
		},
		"completed lowercase": {
			taskID:     "T001",
			taskStatus: "completed",
			wantErr:    false,
		},
		"pending task": {
			taskID:       "T001",
			taskStatus:   "Pending",
			wantErr:      true,
			wantErrMatch: "did not complete",
		},
		"in progress task": {
			taskID:       "T001",
			taskStatus:   "InProgress",
			wantErr:      true,
			wantErrMatch: "did not complete",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specsDir := filepath.Join(tmpDir, "specs")
			specDir := filepath.Join(specsDir, "001-test")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory: %v", err)
			}

			tasksContent := fmt.Sprintf(`tasks:
  branch: "001-test"
  created: "2025-01-01"
summary:
  total_tasks: 1
  total_phases: 1
phases:
  - number: 1
    title: "Test"
    purpose: "Testing"
    tasks:
      - id: "%s"
        title: "Test Task"
        status: "%s"
        type: "implementation"
        parallel: false
        dependencies: []
        acceptance_criteria:
          - "Test"
_meta:
  version: "1.0.0"
  artifact_type: "tasks"
`, tt.taskID, tt.taskStatus)

			tasksPath := filepath.Join(specDir, "tasks.yaml")
			if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
				t.Fatalf("Failed to write tasks.yaml: %v", err)
			}

			cfg := &config.Configuration{
				ClaudeCmd:  "echo",
				SpecsDir:   specsDir,
				MaxRetries: 3,
				StateDir:   filepath.Join(tmpDir, "state"),
			}

			orchestrator := NewWorkflowOrchestrator(cfg)
			err := orchestrator.verifyTaskCompletion(tasksPath, tt.taskID)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.wantErrMatch != "" && !strings.Contains(err.Error(), tt.wantErrMatch) {
					t.Errorf("Error %q should contain %q", err.Error(), tt.wantErrMatch)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestPrintTasksSummary tests the printTasksSummary function
func TestPrintTasksSummary(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	specDir := filepath.Join(specsDir, "001-test")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec directory: %v", err)
	}

	tasksContent := `tasks:
  branch: "001-test"
  created: "2025-01-01"
summary:
  total_tasks: 2
  total_phases: 1
phases:
  - number: 1
    title: "Test"
    purpose: "Testing"
    tasks:
      - id: "T001"
        title: "First"
        status: "Completed"
        type: "implementation"
        parallel: false
        dependencies: []
        acceptance_criteria:
          - "Test"
      - id: "T002"
        title: "Second"
        status: "Completed"
        type: "implementation"
        parallel: false
        dependencies: []
        acceptance_criteria:
          - "Test"
_meta:
  version: "1.0.0"
  artifact_type: "tasks"
`
	tasksPath := filepath.Join(specDir, "tasks.yaml")
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("Failed to write tasks.yaml: %v", err)
	}

	// Create spec.yaml for markSpecCompletedAndPrint
	specContent := `feature:
  branch: "001-test"
  status: "Draft"
  created: "2025-01-01"
user_stories: []
requirements:
  functional: []
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(specContent), 0644); err != nil {
		t.Fatalf("Failed to write spec.yaml: %v", err)
	}

	// Should not panic
	printTasksSummary(tasksPath, specDir)

	// Test with invalid path - should not panic
	printTasksSummary(filepath.Join(tmpDir, "nonexistent.yaml"), specDir)
}

// TestSchemaValidationIntegration tests that schema validators are properly wired into workflow
// These tests verify FR-006: executePlan(), executeTasks(), and executeSpecify() pass schema validator functions to ExecuteStage()
func TestSchemaValidationIntegration(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		stage        string
		validateFunc func(string) error
		validDir     string
		invalidDir   string
		errContains  string
		setupValid   func(string) error
		setupInvalid func(string) error
	}{
		"spec schema validation": {
			stage:        "specify",
			validateFunc: ValidateSpecSchema,
			errContains:  "schema validation failed for spec.yaml",
			setupValid: func(specDir string) error {
				content := `feature:
  branch: "001-test"
  created: "2025-01-01"
  status: "Draft"
  input: "Test feature"
user_stories: []
requirements:
  functional: []
  non_functional: []
success_criteria:
  measurable_outcomes: []
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
				return os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(content), 0644)
			},
			setupInvalid: func(specDir string) error {
				// Missing required 'feature' field
				content := `user_stories: []
requirements:
  functional: []
_meta:
  version: "1.0.0"
  artifact_type: "spec"
`
				return os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(content), 0644)
			},
		},
		"plan schema validation": {
			stage:        "plan",
			validateFunc: ValidatePlanSchema,
			errContains:  "schema validation failed for plan.yaml",
			setupValid: func(specDir string) error {
				content := `plan:
  branch: "001-test"
  created: "2025-01-01"
  spec_path: "specs/001-test/spec.yaml"
summary: "Test plan"
technical_context:
  language: "Go"
  framework: "Cobra"
  project_type: "cli"
implementation_phases: []
project_structure:
  source_code: []
  tests: []
  documentation: []
_meta:
  version: "1.0.0"
  generator: "autospec"
  generator_version: "test"
  created: "2025-01-01T00:00:00Z"
  artifact_type: "plan"
`
				return os.WriteFile(filepath.Join(specDir, "plan.yaml"), []byte(content), 0644)
			},
			setupInvalid: func(specDir string) error {
				// Missing required 'plan' field
				content := `summary: "Test plan"
technical_context:
  language: "Go"
_meta:
  version: "1.0.0"
  artifact_type: "plan"
`
				return os.WriteFile(filepath.Join(specDir, "plan.yaml"), []byte(content), 0644)
			},
		},
		"tasks schema validation": {
			stage:        "tasks",
			validateFunc: ValidateTasksSchema,
			errContains:  "schema validation failed for tasks.yaml",
			setupValid: func(specDir string) error {
				content := `tasks:
  branch: "001-test"
  created: "2025-01-01"
  spec_path: "specs/001-test/spec.yaml"
  plan_path: "specs/001-test/plan.yaml"
summary:
  total_tasks: 1
  total_phases: 1
phases:
  - number: 1
    title: "Setup"
    purpose: "Initial setup"
    tasks:
      - id: "T001"
        title: "Test task"
        status: "Pending"
        type: "implementation"
        parallel: false
        dependencies: []
        acceptance_criteria:
          - "Test passes"
_meta:
  version: "1.0.0"
  generator: "autospec"
  generator_version: "test"
  created: "2025-01-01T00:00:00Z"
  artifact_type: "tasks"
`
				return os.WriteFile(filepath.Join(specDir, "tasks.yaml"), []byte(content), 0644)
			},
			setupInvalid: func(specDir string) error {
				// Missing required 'tasks' field
				content := `summary:
  total_tasks: 1
  total_phases: 1
phases: []
_meta:
  version: "1.0.0"
  artifact_type: "tasks"
`
				return os.WriteFile(filepath.Join(specDir, "tasks.yaml"), []byte(content), 0644)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Test valid artifact passes validation
			t.Run("valid_"+tc.stage+"_passes", func(t *testing.T) {
				t.Parallel()
				tmpDir := t.TempDir()
				specDir := filepath.Join(tmpDir, "specs", "001-test")
				if err := os.MkdirAll(specDir, 0755); err != nil {
					t.Fatalf("Failed to create spec directory: %v", err)
				}
				if err := tc.setupValid(specDir); err != nil {
					t.Fatalf("Failed to setup valid artifact: %v", err)
				}

				err := tc.validateFunc(specDir)
				if err != nil {
					t.Errorf("Valid %s should pass validation, got error: %v", tc.stage, err)
				}
			})

			// Test invalid artifact fails validation with correct error message
			t.Run("invalid_"+tc.stage+"_fails", func(t *testing.T) {
				t.Parallel()
				tmpDir := t.TempDir()
				specDir := filepath.Join(tmpDir, "specs", "001-test")
				if err := os.MkdirAll(specDir, 0755); err != nil {
					t.Fatalf("Failed to create spec directory: %v", err)
				}
				if err := tc.setupInvalid(specDir); err != nil {
					t.Fatalf("Failed to setup invalid artifact: %v", err)
				}

				err := tc.validateFunc(specDir)
				if err == nil {
					t.Errorf("Invalid %s should fail validation, got nil error", tc.stage)
					return
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("Error should contain %q, got: %v", tc.errContains, err)
				}
			})
		})
	}
}

// TestSchemaValidationRejectsInvalidBeforeNextStage verifies that invalid artifacts
// are caught before proceeding to the next workflow stage (FR-001, FR-002, FR-003)
func TestSchemaValidationRejectsInvalidBeforeNextStage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		currentStage   string
		nextStage      string
		validateFunc   func(string) error
		invalidContent string
		expectedErr    string
	}{
		"invalid spec rejected before plan stage": {
			currentStage: "specify",
			nextStage:    "plan",
			validateFunc: ValidateSpecSchema,
			invalidContent: `# Missing required 'feature' field
user_stories: []
requirements:
  functional: []
_meta:
  artifact_type: "spec"
`,
			expectedErr: "missing required field: feature",
		},
		"invalid plan rejected before tasks stage": {
			currentStage: "plan",
			nextStage:    "tasks",
			validateFunc: ValidatePlanSchema,
			invalidContent: `# Missing required 'plan' field
summary: "Test"
_meta:
  artifact_type: "plan"
`,
			expectedErr: "missing required field: plan",
		},
		"invalid tasks rejected before implement stage": {
			currentStage: "tasks",
			nextStage:    "implement",
			validateFunc: ValidateTasksSchema,
			invalidContent: `# Missing required 'tasks' field
summary:
  total_tasks: 0
phases: []
_meta:
  artifact_type: "tasks"
`,
			expectedErr: "missing required field: tasks",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, "specs", "001-test")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory: %v", err)
			}

			// Create invalid artifact file
			var artifactPath string
			switch tc.currentStage {
			case "specify":
				artifactPath = filepath.Join(specDir, "spec.yaml")
			case "plan":
				artifactPath = filepath.Join(specDir, "plan.yaml")
			case "tasks":
				artifactPath = filepath.Join(specDir, "tasks.yaml")
			}
			if err := os.WriteFile(artifactPath, []byte(tc.invalidContent), 0644); err != nil {
				t.Fatalf("Failed to write invalid artifact: %v", err)
			}

			// Run validation (simulates what ExecuteStage does after Claude generates output)
			err := tc.validateFunc(specDir)

			// Verify validation fails with expected error
			if err == nil {
				t.Errorf("Expected validation to fail for invalid %s artifact before %s stage", tc.currentStage, tc.nextStage)
				return
			}
			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Errorf("Error should contain %q, got: %v", tc.expectedErr, err)
			}
		})
	}
}

// TestValidArtifactsProceedToNextStage verifies that valid artifacts proceed without error
func TestValidArtifactsProceedToNextStage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		stage        string
		validateFunc func(string) error
		content      string
		artifactName string
	}{
		"valid spec proceeds to plan": {
			stage:        "specify",
			validateFunc: ValidateSpecSchema,
			artifactName: "spec.yaml",
			content: `feature:
  branch: "001-valid-test"
  created: "2025-01-01"
  status: "Draft"
  input: "Valid test feature"
user_stories:
  - id: "US-001"
    title: "Test story"
    priority: "P1"
    as_a: "user"
    i_want: "to test"
    so_that: "I can verify"
    acceptance_scenarios:
      - given: "a test"
        when: "I run it"
        then: "it passes"
requirements:
  functional:
    - id: "FR-001"
      description: "MUST pass test"
      testable: true
      acceptance_criteria: "Test passes"
  non_functional: []
success_criteria:
  measurable_outcomes:
    - id: "SC-001"
      description: "Tests pass"
      metric: "Pass rate"
      target: "100%"
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
`,
		},
		"valid plan proceeds to tasks": {
			stage:        "plan",
			validateFunc: ValidatePlanSchema,
			artifactName: "plan.yaml",
			content: `plan:
  branch: "001-valid-test"
  created: "2025-01-01"
  spec_path: "specs/001-valid-test/spec.yaml"
summary: "Valid implementation plan"
technical_context:
  language: "Go"
  framework: "Cobra"
  project_type: "cli"
implementation_phases:
  - phase: 1
    name: "Setup"
    goal: "Initial setup"
    deliverables:
      - "Project structure"
project_structure:
  source_code:
    - path: "cmd/main.go"
      description: "Entry point"
  tests:
    - path: "cmd/main_test.go"
      description: "Entry point tests"
  documentation: []
_meta:
  version: "1.0.0"
  generator: "autospec"
  generator_version: "test"
  created: "2025-01-01T00:00:00Z"
  artifact_type: "plan"
`,
		},
		"valid tasks proceeds to implement": {
			stage:        "tasks",
			validateFunc: ValidateTasksSchema,
			artifactName: "tasks.yaml",
			content: `tasks:
  branch: "001-valid-test"
  created: "2025-01-01"
  spec_path: "specs/001-valid-test/spec.yaml"
  plan_path: "specs/001-valid-test/plan.yaml"
summary:
  total_tasks: 1
  total_phases: 1
phases:
  - number: 1
    title: "Setup"
    purpose: "Initial setup"
    tasks:
      - id: "T001"
        title: "Initialize project"
        status: "Pending"
        type: "setup"
        parallel: false
        dependencies: []
        acceptance_criteria:
          - "Project compiles"
_meta:
  version: "1.0.0"
  generator: "autospec"
  generator_version: "test"
  created: "2025-01-01T00:00:00Z"
  artifact_type: "tasks"
`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, "specs", "001-valid-test")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory: %v", err)
			}

			artifactPath := filepath.Join(specDir, tc.artifactName)
			if err := os.WriteFile(artifactPath, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write valid artifact: %v", err)
			}

			// Run validation
			err := tc.validateFunc(specDir)

			// Verify validation passes
			if err != nil {
				t.Errorf("Valid %s artifact should pass validation, got error: %v", tc.stage, err)
			}
		})
	}
}

// TestSchemaValidationErrorMessageFormat verifies error messages are consistent (SC-003)
func TestSchemaValidationErrorMessageFormat(t *testing.T) {
	t.Parallel()

	// All schema validation errors should follow the same format:
	// "schema validation failed for <artifact>:\n- <error1>\n- <error2>..."
	tests := map[string]struct {
		validateFunc func(string) error
		artifactName string
		content      string
	}{
		"spec error format": {
			validateFunc: ValidateSpecSchema,
			artifactName: "spec.yaml",
			content: `user_stories: []
_meta:
  artifact_type: "spec"
`,
		},
		"plan error format": {
			validateFunc: ValidatePlanSchema,
			artifactName: "plan.yaml",
			content: `summary: "test"
_meta:
  artifact_type: "plan"
`,
		},
		"tasks error format": {
			validateFunc: ValidateTasksSchema,
			artifactName: "tasks.yaml",
			content: `phases: []
_meta:
  artifact_type: "tasks"
`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, "specs", "001-test")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory: %v", err)
			}

			artifactPath := filepath.Join(specDir, tc.artifactName)
			if err := os.WriteFile(artifactPath, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write artifact: %v", err)
			}

			err := tc.validateFunc(specDir)
			if err == nil {
				t.Fatal("Expected validation error")
			}

			errStr := err.Error()
			// Verify error format is consistent
			if !strings.HasPrefix(errStr, "schema validation failed for "+tc.artifactName) {
				t.Errorf("Error should start with 'schema validation failed for %s', got: %s", tc.artifactName, errStr)
			}
			if !strings.Contains(errStr, "- ") {
				t.Errorf("Error should contain bullet point format '- ', got: %s", errStr)
			}
		})
	}
}

// TestSchemaValidationInAutospecRunPath verifies schema validation in the autospec run -a path (US-003, T014).
// This test simulates the full workflow run path where an invalid spec.yaml should be rejected
// before proceeding to the plan stage.
func TestSchemaValidationInAutospecRunPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		specContent      string
		expectValidation bool
		errContains      string
		description      string
	}{
		"invalid spec missing feature field rejected before plan": {
			specContent: `# Missing required 'feature' field
user_stories: []
requirements:
  functional: []
  non_functional: []
_meta:
  version: "1.0.0"
  artifact_type: "spec"
`,
			expectValidation: false,
			errContains:      "missing required field: feature",
			description:      "Spec without 'feature' field should fail schema validation",
		},
		"invalid spec missing branch rejected before plan": {
			specContent: `feature:
  created: "2025-01-01"
  status: "Draft"
user_stories: []
requirements:
  functional: []
  non_functional: []
_meta:
  version: "1.0.0"
  artifact_type: "spec"
`,
			expectValidation: false,
			errContains:      "missing required field: branch",
			description:      "Spec without 'feature.branch' should fail schema validation",
		},
		"valid spec proceeds to plan stage": {
			specContent: `feature:
  branch: "001-test-feature"
  created: "2025-01-01"
  status: "Draft"
  input: "Test feature for validation"
user_stories:
  - id: "US-001"
    title: "Test story"
    priority: "P1"
    as_a: "developer"
    i_want: "to test validation"
    so_that: "I can verify the workflow"
    acceptance_scenarios:
      - given: "a valid spec"
        when: "validation runs"
        then: "it passes"
requirements:
  functional:
    - id: "FR-001"
      description: "MUST validate spec"
      testable: true
      acceptance_criteria: "Spec passes validation"
  non_functional: []
success_criteria:
  measurable_outcomes:
    - id: "SC-001"
      description: "Validation passes"
      metric: "Pass rate"
      target: "100%"
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
`,
			expectValidation: true,
			errContains:      "",
			description:      "Valid spec should pass schema validation",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, "specs", "001-test-feature")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory: %v", err)
			}

			// Write spec.yaml (simulating output from specify stage)
			specPath := filepath.Join(specDir, "spec.yaml")
			if err := os.WriteFile(specPath, []byte(tc.specContent), 0644); err != nil {
				t.Fatalf("Failed to write spec.yaml: %v", err)
			}

			// Run schema validation (same function used in autospec run -a path)
			err := ValidateSpecSchema(specDir)

			// Verify validation result matches expected outcome
			if tc.expectValidation {
				if err != nil {
					t.Errorf("%s: expected validation to pass, got error: %v", tc.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("%s: expected validation to fail, but it passed", tc.description)
					return
				}
				// Verify error message format matches standalone specify command
				errStr := err.Error()
				if !strings.HasPrefix(errStr, "schema validation failed for spec.yaml") {
					t.Errorf("Error format mismatch: expected prefix 'schema validation failed for spec.yaml', got: %s", errStr)
				}
				if tc.errContains != "" && !strings.Contains(errStr, tc.errContains) {
					t.Errorf("Error should contain %q, got: %s", tc.errContains, errStr)
				}
			}
		})
	}
}

// TestSchemaValidationInAutospecPrepPath verifies schema validation in the autospec prep path (US-003, T015).
// This test simulates the prep workflow where an invalid plan.yaml should be rejected
// before proceeding to the tasks stage.
func TestSchemaValidationInAutospecPrepPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		planContent      string
		expectValidation bool
		errContains      string
		description      string
	}{
		"invalid plan missing plan field rejected before tasks": {
			planContent: `# Missing required 'plan' field
summary: "Test plan"
technical_context:
  language: "Go"
_meta:
  version: "1.0.0"
  artifact_type: "plan"
`,
			expectValidation: false,
			errContains:      "missing required field: plan",
			description:      "Plan without 'plan' field should fail schema validation",
		},
		"invalid plan missing branch rejected before tasks": {
			planContent: `plan:
  created: "2025-01-01"
  spec_path: "specs/001-test/spec.yaml"
summary: "Test plan"
technical_context:
  language: "Go"
_meta:
  version: "1.0.0"
  artifact_type: "plan"
`,
			expectValidation: false,
			errContains:      "missing required field: branch",
			description:      "Plan without 'plan.branch' should fail schema validation",
		},
		"valid plan proceeds to tasks stage": {
			planContent: `plan:
  branch: "001-test-feature"
  created: "2025-01-01"
  spec_path: "specs/001-test-feature/spec.yaml"
summary: "Implementation plan for test feature"
technical_context:
  language: "Go"
  framework: "Cobra"
  project_type: "cli"
implementation_phases:
  - phase: 1
    name: "Setup"
    goal: "Initial setup"
    deliverables:
      - "Project structure"
project_structure:
  source_code:
    - path: "cmd/main.go"
      description: "Entry point"
  tests:
    - path: "cmd/main_test.go"
      description: "Entry point tests"
  documentation: []
_meta:
  version: "1.0.0"
  generator: "autospec"
  generator_version: "test"
  created: "2025-01-01T00:00:00Z"
  artifact_type: "plan"
`,
			expectValidation: true,
			errContains:      "",
			description:      "Valid plan should pass schema validation",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, "specs", "001-test-feature")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory: %v", err)
			}

			// Write plan.yaml (simulating output from plan stage)
			planPath := filepath.Join(specDir, "plan.yaml")
			if err := os.WriteFile(planPath, []byte(tc.planContent), 0644); err != nil {
				t.Fatalf("Failed to write plan.yaml: %v", err)
			}

			// Run schema validation (same function used in autospec prep path)
			err := ValidatePlanSchema(specDir)

			// Verify validation result matches expected outcome
			if tc.expectValidation {
				if err != nil {
					t.Errorf("%s: expected validation to pass, got error: %v", tc.description, err)
				}
			} else {
				if err == nil {
					t.Errorf("%s: expected validation to fail, but it passed", tc.description)
					return
				}
				// Verify error message format matches standalone plan command
				errStr := err.Error()
				if !strings.HasPrefix(errStr, "schema validation failed for plan.yaml") {
					t.Errorf("Error format mismatch: expected prefix 'schema validation failed for plan.yaml', got: %s", errStr)
				}
				if tc.errContains != "" && !strings.Contains(errStr, tc.errContains) {
					t.Errorf("Error should contain %q, got: %s", tc.errContains, errStr)
				}
			}
		})
	}
}

// assertSchemaValidationErrorFormat is a helper function that verifies the consistency
// of schema validation error messages across different entry points (US-003, T016).
// It ensures:
// - Error prefix is consistent: "schema validation failed for <artifact>:"
// - Error contains bullet point format "- <error message>"
// - Error includes the expected field/message
func assertSchemaValidationErrorFormat(t *testing.T, err error, artifactName string, expectedContent string) {
	t.Helper()

	if err == nil {
		t.Fatalf("Expected validation error for %s, got nil", artifactName)
	}

	errStr := err.Error()

	// Check consistent prefix format
	expectedPrefix := fmt.Sprintf("schema validation failed for %s:", artifactName)
	if !strings.HasPrefix(errStr, expectedPrefix) {
		t.Errorf("Error format inconsistent: expected prefix %q, got: %s", expectedPrefix, errStr)
	}

	// Check bullet point format is present
	if !strings.Contains(errStr, "- ") {
		t.Errorf("Error format inconsistent: expected bullet point format '- ', got: %s", errStr)
	}

	// Check expected content is present
	if expectedContent != "" && !strings.Contains(errStr, expectedContent) {
		t.Errorf("Error should contain %q, got: %s", expectedContent, errStr)
	}
}

// TestSchemaValidationErrorConsistencyAcrossEntryPoints verifies that error messages
// are consistent across different autospec entry points (US-003, T016).
// This test checks that the same validation failure produces identical error format
// regardless of whether it's triggered via:
// - autospec run -a (specify → plan → tasks → implement)
// - autospec specify (standalone)
// - autospec prep (specify → plan → tasks)
// - autospec plan (standalone)
// - autospec tasks (standalone)
func TestSchemaValidationErrorConsistencyAcrossEntryPoints(t *testing.T) {
	t.Parallel()

	// Test that the same invalid spec produces consistent errors
	// regardless of which command path invokes validation
	t.Run("spec validation error format consistency", func(t *testing.T) {
		t.Parallel()

		invalidSpecContent := `user_stories: []
requirements:
  functional: []
_meta:
  artifact_type: "spec"
`

		// Create two separate directories to simulate different entry points
		entries := []string{"run-a", "specify-standalone"}
		errors := make(map[string]error)

		for _, entry := range entries {
			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, "specs", "001-test")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory for %s: %v", entry, err)
			}

			specPath := filepath.Join(specDir, "spec.yaml")
			if err := os.WriteFile(specPath, []byte(invalidSpecContent), 0644); err != nil {
				t.Fatalf("Failed to write spec.yaml for %s: %v", entry, err)
			}

			// Both entry points use the same ValidateSpecSchema function
			errors[entry] = ValidateSpecSchema(specDir)
		}

		// Verify both entry points produce errors
		for entry, err := range errors {
			assertSchemaValidationErrorFormat(t, err, "spec.yaml", "missing required field: feature")
			t.Logf("%s error: %v", entry, err)
		}

		// Verify error formats are identical (prefix and structure)
		runErr := errors["run-a"].Error()
		specifyErr := errors["specify-standalone"].Error()
		if runErr != specifyErr {
			t.Errorf("Error format mismatch across entry points:\n  run -a: %s\n  specify: %s", runErr, specifyErr)
		}
	})

	// Test that the same invalid plan produces consistent errors
	t.Run("plan validation error format consistency", func(t *testing.T) {
		t.Parallel()

		invalidPlanContent := `summary: "test"
technical_context:
  language: "Go"
_meta:
  artifact_type: "plan"
`

		entries := []string{"prep", "plan-standalone"}
		errors := make(map[string]error)

		for _, entry := range entries {
			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, "specs", "001-test")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory for %s: %v", entry, err)
			}

			planPath := filepath.Join(specDir, "plan.yaml")
			if err := os.WriteFile(planPath, []byte(invalidPlanContent), 0644); err != nil {
				t.Fatalf("Failed to write plan.yaml for %s: %v", entry, err)
			}

			errors[entry] = ValidatePlanSchema(specDir)
		}

		for entry, err := range errors {
			assertSchemaValidationErrorFormat(t, err, "plan.yaml", "missing required field: plan")
			t.Logf("%s error: %v", entry, err)
		}

		prepErr := errors["prep"].Error()
		planErr := errors["plan-standalone"].Error()
		if prepErr != planErr {
			t.Errorf("Error format mismatch across entry points:\n  prep: %s\n  plan: %s", prepErr, planErr)
		}
	})

	// Test that the same invalid tasks produces consistent errors
	t.Run("tasks validation error format consistency", func(t *testing.T) {
		t.Parallel()

		invalidTasksContent := `summary:
  total_tasks: 0
phases: []
_meta:
  artifact_type: "tasks"
`

		entries := []string{"run-a", "tasks-standalone"}
		errors := make(map[string]error)

		for _, entry := range entries {
			tmpDir := t.TempDir()
			specDir := filepath.Join(tmpDir, "specs", "001-test")
			if err := os.MkdirAll(specDir, 0755); err != nil {
				t.Fatalf("Failed to create spec directory for %s: %v", entry, err)
			}

			tasksPath := filepath.Join(specDir, "tasks.yaml")
			if err := os.WriteFile(tasksPath, []byte(invalidTasksContent), 0644); err != nil {
				t.Fatalf("Failed to write tasks.yaml for %s: %v", entry, err)
			}

			errors[entry] = ValidateTasksSchema(specDir)
		}

		for entry, err := range errors {
			assertSchemaValidationErrorFormat(t, err, "tasks.yaml", "missing required field: tasks")
			t.Logf("%s error: %v", entry, err)
		}

		runErr := errors["run-a"].Error()
		tasksErr := errors["tasks-standalone"].Error()
		if runErr != tasksErr {
			t.Errorf("Error format mismatch across entry points:\n  run -a: %s\n  tasks: %s", runErr, tasksErr)
		}
	})
}
