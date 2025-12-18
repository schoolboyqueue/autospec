# Arch 4: Complete Dependency Injection (REVISED)

**Location:** Multiple packages (workflow, config, validation, retry)
**Impact:** MEDIUM - Completes testability improvements started in arch-1/2
**Effort:** LOW-MEDIUM (reduced scope due to prior work)
**Dependencies:**
- arch-1 and arch-2 COMPLETED (interfaces already exist)
- arch-3 IN PROGRESS (CLI subpackages - coordinate deps wiring)

## Revision Context

**Original advice was:** Do arch-4 BEFORE arch-1 and arch-2.

**What actually happened:** arch-1 and arch-2 were completed first, creating interfaces organically as needed.

**Why original order was recommended:**
- Define all interfaces upfront → subsequent refactoring uses them naturally
- Avoids interface discovery during refactoring
- Creates consistent abstraction layer from the start

**Current situation:** Most executor interfaces already exist. This task now focuses on:
1. Remaining gaps (Configuration, Validator, RetryStore interfaces)
2. Consolidating the deps pattern in WorkflowOrchestrator
3. Coordinating with arch-3's CLI subpackages for dependency wiring

---

## Already Completed (from arch-1/arch-2)

### Interfaces in `internal/workflow/interfaces.go`

```go
// ✓ ClaudeRunner - abstracts Claude command execution
type ClaudeRunner interface {
    Execute(prompt string) error
    FormatCommand(prompt string) string
}

// ✓ StageExecutorInterface - specify, plan, tasks stages
type StageExecutorInterface interface {
    ExecuteSpecify(featureDescription string) (string, error)
    ExecutePlan(specNameArg string, prompt string) error
    ExecuteTasks(specNameArg string, prompt string) error
    ExecuteConstitution(prompt string) error
    ExecuteClarify(specName string, prompt string) error
    ExecuteChecklist(specName string, prompt string) error
    ExecuteAnalyze(specName string, prompt string) error
}

// ✓ PhaseExecutorInterface - phase-based implementation
type PhaseExecutorInterface interface {
    ExecutePhaseLoop(...) error
    ExecuteSinglePhase(...) error
    ExecuteDefault(...) error
}

// ✓ TaskExecutorInterface - task-level execution
type TaskExecutorInterface interface {
    ExecuteTaskLoop(...) error
    ExecuteSingleTask(...) error
    PrepareTaskExecution(...) ([]validation.TaskItem, int, int, error)
}
```

### Extracted Components (from arch-2)

- `ProgressController` in `progress_controller.go` - display concerns
- `NotifyDispatcher` in `notify_dispatcher.go` - notification routing

### Mock Implementations in `mocks_test.go`

- `MockClaudeExecutor`
- `MockStageExecutor`
- `MockPhaseExecutor`
- `MockTaskExecutor`
- `mockPreflightChecker`

---

## Remaining Gaps

### 1. Configuration Interface (NOT DONE)

**Current state:** WorkflowOrchestrator uses concrete `*config.Configuration`

```go
// Current - concrete type
type WorkflowOrchestrator struct {
    Config *config.Configuration  // Concrete, hard to mock
    // ...
}
```

**Target:** Interface for config access

```go
// Define in internal/workflow/interfaces.go
type ConfigProvider interface {
    GetClaudeCmd() string
    GetCustomClaudeCmd() string
    GetClaudeArgs() []string
    GetMaxRetries() int
    GetTimeout() int
    GetSpecsDir() string
    GetStateDir() string
    GetImplementMethod() string
    GetNotifications() notify.NotificationConfig
    IsSkipPreflight() bool
    IsSkipConfirmations() bool
}

// Adapter in internal/config/adapter.go
func (c *Configuration) GetClaudeCmd() string       { return c.ClaudeCmd }
func (c *Configuration) GetMaxRetries() int          { return c.MaxRetries }
// ... etc
```

### 2. Validator Interface (NOT DONE)

**Current state:** Direct function calls to validation package

```go
// Current - direct function calls
result := validation.ValidateSpec(specPath)
phases, err := validation.ExtractPhaseInfo(tasksPath)
```

**Target:** Interface for validation operations

```go
// Define in internal/workflow/interfaces.go
type ArtifactValidator interface {
    ValidateSpec(specPath string) *validation.Result
    ValidatePlan(planPath string) *validation.Result
    ValidateTasks(tasksPath string) *validation.Result
    ExtractPhaseInfo(tasksPath string) ([]validation.PhaseInfo, error)
    ExtractOrderedTasks(tasksPath string) ([]validation.TaskItem, error)
}

// Default implementation wraps existing functions
type defaultValidator struct{}

func (v *defaultValidator) ValidateSpec(path string) *validation.Result {
    return validation.ValidateSpec(path)
}
```

### 3. RetryStore Interface (NOT DONE)

**Current state:** Package-level functions with stateDir parameter

```go
// Current - package functions
state, err := retry.LoadRetryState(stateDir, specName, phase, maxRetries)
err := retry.SaveRetryState(stateDir, state)
```

**Target:** Interface-based retry store

```go
// Define in internal/workflow/interfaces.go
type RetryStateStore interface {
    LoadState(specName, phase string, maxRetries int) (*retry.RetryState, error)
    SaveState(state *retry.RetryState) error
    LoadStageState(specName string) (*retry.StageExecutionState, error)
    SaveStageState(state *retry.StageExecutionState) error
    LoadTaskState(specName string) (*retry.TaskExecutionState, error)
    SaveTaskState(state *retry.TaskExecutionState) error
    ClearAll(specName string) error
}

// Implementation in internal/retry/store.go
type FileStore struct {
    StateDir string
}

func NewFileStore(stateDir string) *FileStore {
    return &FileStore{StateDir: stateDir}
}

func (s *FileStore) LoadState(specName, phase string, maxRetries int) (*RetryState, error) {
    return LoadRetryState(s.StateDir, specName, phase, maxRetries)
}
```

---

## Target Pattern: Deps Struct

### WorkflowOrchestrator Deps

```go
// internal/workflow/orchestrator.go

type OrchestratorDeps struct {
    Config        ConfigProvider           // Interface
    StageExecutor StageExecutorInterface   // Already interface
    PhaseExecutor PhaseExecutorInterface   // Already interface
    TaskExecutor  TaskExecutorInterface    // Already interface
    Validator     ArtifactValidator        // New interface
    RetryStore    RetryStateStore          // New interface
    Preflight     PreflightChecker         // Already interface
}

func NewWorkflowOrchestrator(deps OrchestratorDeps) *WorkflowOrchestrator {
    return &WorkflowOrchestrator{
        config:        deps.Config,
        stageExecutor: deps.StageExecutor,
        phaseExecutor: deps.PhaseExecutor,
        taskExecutor:  deps.TaskExecutor,
        validator:     deps.Validator,
        retryStore:    deps.RetryStore,
        preflight:     deps.Preflight,
    }
}
```

### Executor Deps (for Stage/Phase/Task executors)

```go
type ExecutorDeps struct {
    Claude     ClaudeRunner      // Already interface
    Progress   *ProgressController
    Notify     *NotifyDispatcher
    Validator  ArtifactValidator // New interface
    RetryStore RetryStateStore   // New interface
}
```

---

## Integration with Arch-3 (CLI Subpackages)

Arch-3 is creating CLI subpackages (`stages/`, `config/`, `util/`, `admin/`).

### Dependency Wiring Location

Dependencies should be wired in the CLI layer, with a factory pattern:

```go
// internal/cli/factory.go (or in root.go)

type CLIDeps struct {
    Config     *config.Configuration
    Validator  workflow.ArtifactValidator
    RetryStore workflow.RetryStateStore
    Notifier   *notify.Handler
}

func NewCLIDeps(cfg *config.Configuration) *CLIDeps {
    return &CLIDeps{
        Config:     cfg,
        Validator:  workflow.NewDefaultValidator(),
        RetryStore: retry.NewFileStore(cfg.StateDir),
        Notifier:   notify.NewHandler(cfg.Notifications),
    }
}

func (d *CLIDeps) NewOrchestrator() *workflow.WorkflowOrchestrator {
    return workflow.NewWorkflowOrchestrator(workflow.OrchestratorDeps{
        Config:     d.Config,
        Validator:  d.Validator,
        RetryStore: d.RetryStore,
        // ... wire up executors
    })
}
```

### Subpackage Coordination

Each CLI subpackage receives deps via parent:

```go
// internal/cli/stages/register.go
func Register(parent *cobra.Command, deps *cli.CLIDeps) {
    parent.AddCommand(newSpecifyCmd(deps))
    parent.AddCommand(newPlanCmd(deps))
    // ...
}
```

---

## Implementation Approach

### Phase 1: Define Remaining Interfaces (workflow package)

1. Add `ConfigProvider` interface to `interfaces.go`
2. Add `ArtifactValidator` interface to `interfaces.go`
3. Add `RetryStateStore` interface to `interfaces.go`

### Phase 2: Create Implementations

4. Add adapter methods to `config.Configuration` (implements ConfigProvider)
5. Create `defaultValidator` struct (implements ArtifactValidator)
6. Create `retry.FileStore` struct (implements RetryStateStore)

### Phase 3: Update Constructors

7. Update executor constructors to accept interface deps
8. Update WorkflowOrchestrator to use OrchestratorDeps pattern
9. Create CLI factory for dependency wiring

### Phase 4: Mock Implementations

10. Add `MockConfigProvider` to `mocks_test.go`
11. Add `MockValidator` to `mocks_test.go`
12. Add `MockRetryStore` to `mocks_test.go`

### Phase 5: Update Tests

13. Update existing tests to use new mocks
14. Add tests for new interface implementations
15. Verify all tests pass

---

## Acceptance Criteria

### Interfaces (in interfaces.go)
- [ ] ConfigProvider interface defined
- [ ] ArtifactValidator interface defined
- [ ] RetryStateStore interface defined

### Implementations
- [ ] Configuration implements ConfigProvider (via adapter methods)
- [ ] defaultValidator implements ArtifactValidator
- [ ] retry.FileStore implements RetryStateStore

### Dependency Pattern
- [ ] WorkflowOrchestrator uses OrchestratorDeps struct
- [ ] Executors use ExecutorDeps struct
- [ ] CLI layer wires dependencies via factory

### Mocks
- [ ] MockConfigProvider for testing
- [ ] MockValidator for testing
- [ ] MockRetryStore for testing

### Quality Gates
- [ ] All tests pass
- [ ] Build succeeds
- [ ] No breaking changes to CLI commands

---

## Non-Functional Requirements

- Interfaces in dedicated `interfaces.go` files
- Mocks in `mocks_test.go` files
- Accept interfaces, return concrete types
- No breaking changes to existing CLI commands
- Coordinate with arch-3 CLI subpackage structure
- All functions under 40 lines
- All errors wrapped with context

---

## Files to Modify

| File | Changes |
|------|---------|
| `internal/workflow/interfaces.go` | Add ConfigProvider, ArtifactValidator, RetryStateStore |
| `internal/workflow/mocks_test.go` | Add mock implementations |
| `internal/workflow/orchestrator.go` | Use OrchestratorDeps pattern |
| `internal/workflow/validator.go` | New file: defaultValidator impl |
| `internal/config/config.go` | Add ConfigProvider adapter methods |
| `internal/retry/store.go` | New file: FileStore impl |
| `internal/cli/factory.go` | New file: dependency wiring |
| `internal/cli/root.go` | Use factory for orchestrator creation |

---

## Command

```bash
autospec specify "$(cat .dev/tasks/arch/arch-4-dependency-injection.md)"
```
