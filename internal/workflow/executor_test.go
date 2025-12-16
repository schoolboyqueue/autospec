package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ariel-frischer/autospec/internal/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClaudeExecutor implements a mock for testing
type mockClaudeExecutor struct {
	executeErr   error
	executeCalls []string
}

func (m *mockClaudeExecutor) Execute(prompt string) error {
	m.executeCalls = append(m.executeCalls, prompt)
	return m.executeErr
}

func (m *mockClaudeExecutor) FormatCommand(prompt string) string {
	return "claude " + prompt
}

func TestGetStageNumber(t *testing.T) {
	tests := map[string]struct {
		stage Stage
		want  int
	}{
		"constitution stage": {stage: StageConstitution, want: 1},
		"specify stage":      {stage: StageSpecify, want: 2},
		"clarify stage":      {stage: StageClarify, want: 3},
		"plan stage":         {stage: StagePlan, want: 4},
		"tasks stage":        {stage: StageTasks, want: 5},
		"checklist stage":    {stage: StageChecklist, want: 6},
		"analyze stage":      {stage: StageAnalyze, want: 7},
		"implement stage":    {stage: StageImplement, want: 8},
		"unknown stage":      {stage: Stage("unknown"), want: 0},
		"empty stage":        {stage: Stage(""), want: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			executor := &Executor{}
			got := executor.getStageNumber(tc.stage)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBuildStageInfo(t *testing.T) {
	tests := map[string]struct {
		stage       Stage
		retryCount  int
		maxRetries  int
		totalStages int
		wantName    string
		wantNumber  int
	}{
		"specify stage no retries": {
			stage:       StageSpecify,
			retryCount:  0,
			maxRetries:  3,
			totalStages: 4,
			wantName:    "specify",
			wantNumber:  2,
		},
		"plan stage with retries": {
			stage:       StagePlan,
			retryCount:  2,
			maxRetries:  3,
			totalStages: 4,
			wantName:    "plan",
			wantNumber:  4,
		},
		"implement stage max retries": {
			stage:       StageImplement,
			retryCount:  3,
			maxRetries:  3,
			totalStages: 8,
			wantName:    "implement",
			wantNumber:  8,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			executor := &Executor{
				MaxRetries:  tc.maxRetries,
				TotalStages: tc.totalStages,
			}

			info := executor.buildStageInfo(tc.stage, tc.retryCount)

			assert.Equal(t, tc.wantName, info.Name)
			assert.Equal(t, tc.wantNumber, info.Number)
			assert.Equal(t, tc.totalStages, info.TotalStages)
			assert.Equal(t, tc.retryCount, info.RetryCount)
			assert.Equal(t, tc.maxRetries, info.MaxRetries)
		})
	}
}

func TestExecuteStage_Success(t *testing.T) {
	stateDir := t.TempDir()
	specsDir := t.TempDir()

	// Create spec directory with required file
	specDir := filepath.Join(specsDir, "001-test")
	require.NoError(t, os.MkdirAll(specDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# Test Spec"), 0644))

	executor := &Executor{
		Claude: &ClaudeExecutor{
			ClaudeCmd:  "echo",
			ClaudeArgs: []string{},
		},
		StateDir:   stateDir,
		SpecsDir:   specsDir,
		MaxRetries: 3,
	}

	// Validation function that always succeeds
	validateFunc := func(dir string) error {
		return nil
	}

	result, err := executor.ExecuteStage("001-test", StageSpecify, "/test.command", validateFunc)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, StageSpecify, result.Stage)
	assert.Equal(t, 0, result.RetryCount)
	assert.False(t, result.Exhausted)
}

func TestExecuteStage_ValidationFailure(t *testing.T) {
	stateDir := t.TempDir()
	specsDir := t.TempDir()

	executor := &Executor{
		Claude: &ClaudeExecutor{
			ClaudeCmd:  "echo",
			ClaudeArgs: []string{},
		},
		StateDir:   stateDir,
		SpecsDir:   specsDir,
		MaxRetries: 3,
	}

	// Validation function that always fails
	validateFunc := func(dir string) error {
		return errors.New("validation failed: missing spec.md")
	}

	result, err := executor.ExecuteStage("001-test", StageSpecify, "/test.command", validateFunc)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.False(t, result.Success)
	assert.Equal(t, 1, result.RetryCount) // Should have incremented
}

func TestExecuteStage_RetryExhausted(t *testing.T) {
	stateDir := t.TempDir()
	specsDir := t.TempDir()

	// Pre-set retry count to max so next failure returns exhausted error
	state := &retry.RetryState{
		SpecName:   "001-test",
		Phase:      "specify",
		Count:      3,
		MaxRetries: 3,
	}
	require.NoError(t, retry.SaveRetryState(stateDir, state))

	executor := &Executor{
		Claude: &ClaudeExecutor{
			ClaudeCmd:  "echo",
			ClaudeArgs: []string{},
		},
		StateDir:   stateDir,
		SpecsDir:   specsDir,
		MaxRetries: 3,
	}

	// Validation function that always fails
	validateFunc := func(dir string) error {
		return errors.New("validation failed")
	}

	result, err := executor.ExecuteStage("001-test", StageSpecify, "/test.command", validateFunc)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exhausted")
	assert.False(t, result.Success)
	assert.True(t, result.Exhausted)
}

func TestExecuteStage_ResetsRetryOnSuccess(t *testing.T) {
	stateDir := t.TempDir()
	specsDir := t.TempDir()

	// Pre-set retry count
	state := &retry.RetryState{
		SpecName:   "001-test",
		Phase:      "specify",
		Count:      2,
		MaxRetries: 3,
	}
	require.NoError(t, retry.SaveRetryState(stateDir, state))

	executor := &Executor{
		Claude: &ClaudeExecutor{
			ClaudeCmd:  "echo",
			ClaudeArgs: []string{},
		},
		StateDir:   stateDir,
		SpecsDir:   specsDir,
		MaxRetries: 3,
	}

	// Validation function that succeeds
	validateFunc := func(dir string) error {
		return nil
	}

	result, err := executor.ExecuteStage("001-test", StageSpecify, "/test.command", validateFunc)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 0, result.RetryCount)

	// Verify retry state was reset
	loaded, err := retry.LoadRetryState(stateDir, "001-test", "specify", 3)
	require.NoError(t, err)
	assert.Equal(t, 0, loaded.Count)
}

func TestExecuteWithRetry_Success(t *testing.T) {
	executor := &Executor{
		Claude: &ClaudeExecutor{
			ClaudeCmd:  "echo",
			ClaudeArgs: []string{"success"},
		},
	}

	err := executor.ExecuteWithRetry("/test.command", 3)
	assert.NoError(t, err)
}

func TestExecuteWithRetry_AllAttemptsFail(t *testing.T) {
	executor := &Executor{
		Claude: &ClaudeExecutor{
			ClaudeCmd:  "false", // Command that always fails
			ClaudeArgs: []string{},
		},
	}

	err := executor.ExecuteWithRetry("/test.command", 2)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "all 2 attempts failed")
}

func TestGetRetryState(t *testing.T) {
	stateDir := t.TempDir()

	// Save initial state
	state := &retry.RetryState{
		SpecName:   "001-test",
		Phase:      "plan",
		Count:      1,
		MaxRetries: 3,
	}
	require.NoError(t, retry.SaveRetryState(stateDir, state))

	executor := &Executor{
		StateDir:   stateDir,
		MaxRetries: 3,
	}

	loaded, err := executor.GetRetryState("001-test", StagePlan)

	require.NoError(t, err)
	assert.Equal(t, 1, loaded.Count)
	assert.Equal(t, "001-test", loaded.SpecName)
	assert.Equal(t, "plan", loaded.Phase)
}

func TestResetStage(t *testing.T) {
	stateDir := t.TempDir()

	// Save initial state with non-zero count
	state := &retry.RetryState{
		SpecName:   "001-test",
		Phase:      "tasks",
		Count:      2,
		MaxRetries: 3,
	}
	require.NoError(t, retry.SaveRetryState(stateDir, state))

	executor := &Executor{
		StateDir:   stateDir,
		MaxRetries: 3,
	}

	err := executor.ResetStage("001-test", StageTasks)
	require.NoError(t, err)

	// Verify reset
	loaded, err := retry.LoadRetryState(stateDir, "001-test", "tasks", 3)
	require.NoError(t, err)
	assert.Equal(t, 0, loaded.Count)
}

func TestValidateSpec(t *testing.T) {
	t.Run("spec exists", func(t *testing.T) {
		t.Parallel()
		specDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# Spec"), 0644))

		executor := &Executor{}
		err := executor.ValidateSpec(specDir)
		assert.NoError(t, err)
	})

	t.Run("spec missing", func(t *testing.T) {
		t.Parallel()
		specDir := t.TempDir()

		executor := &Executor{}
		err := executor.ValidateSpec(specDir)
		assert.Error(t, err)
	})
}

func TestValidatePlan(t *testing.T) {
	t.Run("plan exists", func(t *testing.T) {
		t.Parallel()
		specDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(specDir, "plan.md"), []byte("# Plan"), 0644))

		executor := &Executor{}
		err := executor.ValidatePlan(specDir)
		assert.NoError(t, err)
	})

	t.Run("plan missing", func(t *testing.T) {
		t.Parallel()
		specDir := t.TempDir()

		executor := &Executor{}
		err := executor.ValidatePlan(specDir)
		assert.Error(t, err)
	})
}

func TestValidateTasks(t *testing.T) {
	t.Run("tasks exists", func(t *testing.T) {
		t.Parallel()
		specDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte("# Tasks"), 0644))

		executor := &Executor{}
		err := executor.ValidateTasks(specDir)
		assert.NoError(t, err)
	})

	t.Run("tasks missing", func(t *testing.T) {
		t.Parallel()
		specDir := t.TempDir()

		executor := &Executor{}
		err := executor.ValidateTasks(specDir)
		assert.Error(t, err)
	})
}

func TestValidateTasksComplete(t *testing.T) {
	tests := map[string]struct {
		content string
		wantErr bool
	}{
		"all tasks complete": {
			content: `# Tasks
- [x] Task 1
- [x] Task 2
- [x] Task 3
`,
			wantErr: false,
		},
		"some tasks incomplete": {
			content: `# Tasks
- [x] Task 1
- [ ] Task 2
- [x] Task 3
`,
			wantErr: true,
		},
		"all tasks incomplete": {
			content: `# Tasks
- [ ] Task 1
- [ ] Task 2
`,
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			tasksPath := filepath.Join(dir, "tasks.md")
			require.NoError(t, os.WriteFile(tasksPath, []byte(tc.content), 0644))

			executor := &Executor{}
			err := executor.ValidateTasksComplete(tasksPath)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "tasks remain")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDebugLog(t *testing.T) {
	t.Run("debug disabled does not print", func(t *testing.T) {
		t.Parallel()
		executor := &Executor{Debug: false}
		// Should not panic and should not print
		executor.debugLog("test message %s", "arg")
	})

	t.Run("debug enabled prints", func(t *testing.T) {
		t.Parallel()
		executor := &Executor{Debug: true}
		// Should not panic - we can't easily capture stdout in this test
		// but we verify it doesn't crash
		executor.debugLog("test message %s", "arg")
	})
}
