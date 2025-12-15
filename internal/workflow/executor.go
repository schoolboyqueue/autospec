package workflow

import (
	"fmt"

	"github.com/ariel-frischer/autospec/internal/progress"
	"github.com/ariel-frischer/autospec/internal/retry"
	"github.com/ariel-frischer/autospec/internal/validation"
)

// Executor handles command execution with retry logic
type Executor struct {
	Claude          *ClaudeExecutor
	StateDir        string
	SpecsDir        string
	MaxRetries      int
	ProgressDisplay *progress.ProgressDisplay // Optional progress display
	TotalPhases     int                       // Total phases in workflow
	Debug           bool                      // Enable debug logging
}

// Phase represents a workflow phase (specify, plan, tasks, implement)
type Phase string

const (
	// Core workflow phases
	PhaseSpecify   Phase = "specify"
	PhasePlan      Phase = "plan"
	PhaseTasks     Phase = "tasks"
	PhaseImplement Phase = "implement"

	// Optional phases
	PhaseConstitution Phase = "constitution"
	PhaseClarify      Phase = "clarify"
	PhaseChecklist    Phase = "checklist"
	PhaseAnalyze      Phase = "analyze"
)

// debugLog prints a debug message if debug mode is enabled
func (e *Executor) debugLog(format string, args ...interface{}) {
	if e.Debug {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// getPhaseNumber returns the sequential number for a phase (1-based)
// For optional phases, this returns their position in the canonical order:
// constitution(1) -> specify(2) -> clarify(3) -> plan(4) -> tasks(5) -> checklist(6) -> analyze(7) -> implement(8)
func (e *Executor) getPhaseNumber(phase Phase) int {
	switch phase {
	case PhaseConstitution:
		return 1
	case PhaseSpecify:
		return 2
	case PhaseClarify:
		return 3
	case PhasePlan:
		return 4
	case PhaseTasks:
		return 5
	case PhaseChecklist:
		return 6
	case PhaseAnalyze:
		return 7
	case PhaseImplement:
		return 8
	default:
		return 0
	}
}

// buildPhaseInfo constructs a PhaseInfo from Phase enum and retry state
func (e *Executor) buildPhaseInfo(phase Phase, retryCount int) progress.PhaseInfo {
	return progress.PhaseInfo{
		Name:        string(phase),
		Number:      e.getPhaseNumber(phase),
		TotalPhases: e.TotalPhases,
		Status:      progress.PhaseInProgress,
		RetryCount:  retryCount,
		MaxRetries:  e.MaxRetries,
	}
}

// PhaseResult represents the result of executing a workflow phase
type PhaseResult struct {
	Phase      Phase
	Success    bool
	Error      error
	RetryCount int
	Exhausted  bool
}

// ExecutePhase executes a workflow phase with validation and retry logic
func (e *Executor) ExecutePhase(specName string, phase Phase, command string, validateFunc func(string) error) (*PhaseResult, error) {
	e.debugLog("ExecutePhase called - spec: %s, phase: %s, command: %s", specName, phase, command)
	result := &PhaseResult{
		Phase:   phase,
		Success: false,
	}

	// Load retry state
	e.debugLog("Loading retry state from: %s", e.StateDir)
	retryState, err := retry.LoadRetryState(e.StateDir, specName, string(phase), e.MaxRetries)
	if err != nil {
		e.debugLog("Failed to load retry state: %v", err)
		return result, fmt.Errorf("failed to load retry state: %w", err)
	}
	e.debugLog("Retry state loaded - count: %d, max: %d", retryState.Count, e.MaxRetries)

	// Build phase info and start progress display
	phaseInfo := e.buildPhaseInfo(phase, retryState.Count)
	if e.ProgressDisplay != nil {
		e.debugLog("Starting progress display")
		if err := e.ProgressDisplay.StartPhase(phaseInfo); err != nil {
			// Log warning but don't fail execution
			fmt.Printf("Warning: progress display error: %v\n", err)
		}
	}

	// Display the full command before execution
	fullCommand := e.Claude.FormatCommand(command)
	fmt.Printf("\n→ Executing: %s\n\n", fullCommand)

	// Execute command
	e.debugLog("About to call Claude.Execute()")
	if err := e.Claude.Execute(command); err != nil {
		e.debugLog("Claude.Execute() returned error: %v", err)
		result.Error = fmt.Errorf("command execution failed: %w", err)

		// Show failure in progress display
		if e.ProgressDisplay != nil {
			e.ProgressDisplay.FailPhase(phaseInfo, result.Error)
		}

		// Increment retry count
		if incrementErr := retryState.Increment(); incrementErr != nil {
			// Check if it's a retry exhausted error
			if exhaustedErr, ok := incrementErr.(*retry.RetryExhaustedError); ok {
				result.Exhausted = true
				result.RetryCount = exhaustedErr.Count

				// Save state before returning
				retry.SaveRetryState(e.StateDir, retryState)

				return result, fmt.Errorf("retry limit exhausted: %w", err)
			}
			return result, incrementErr
		}

		// Save incremented retry state
		if saveErr := retry.SaveRetryState(e.StateDir, retryState); saveErr != nil {
			return result, fmt.Errorf("failed to save retry state: %w", saveErr)
		}

		result.RetryCount = retryState.Count
		return result, result.Error
	}
	e.debugLog("Claude.Execute() completed successfully")

	// Validate output
	specDir := fmt.Sprintf("%s/%s", e.SpecsDir, specName) // Simplified for now
	e.debugLog("Running validation function for spec dir: %s", specDir)
	if err := validateFunc(specDir); err != nil {
		e.debugLog("Validation failed: %v", err)
		result.Error = fmt.Errorf("validation failed: %w", err)

		// Show failure in progress display
		if e.ProgressDisplay != nil {
			e.ProgressDisplay.FailPhase(phaseInfo, result.Error)
		}

		// Increment retry count
		if incrementErr := retryState.Increment(); incrementErr != nil {
			if exhaustedErr, ok := incrementErr.(*retry.RetryExhaustedError); ok {
				result.Exhausted = true
				result.RetryCount = exhaustedErr.Count
				retry.SaveRetryState(e.StateDir, retryState)
				return result, fmt.Errorf("validation failed and retry exhausted: %w", err)
			}
			return result, incrementErr
		}

		retry.SaveRetryState(e.StateDir, retryState)
		result.RetryCount = retryState.Count
		return result, result.Error
	}
	e.debugLog("Validation passed!")

	// Success! Show completion in progress display
	if e.ProgressDisplay != nil {
		e.debugLog("Showing completion in progress display")
		phaseInfo.Status = progress.PhaseCompleted
		if err := e.ProgressDisplay.CompletePhase(phaseInfo); err != nil {
			// Log warning but don't fail execution
			fmt.Printf("Warning: progress display error: %v\n", err)
		}
	}

	// Reset retry count
	e.debugLog("Resetting retry count")
	if err := retry.ResetRetryCount(e.StateDir, specName, string(phase)); err != nil {
		// Log error but don't fail - reset is not critical
		fmt.Printf("Warning: failed to reset retry count: %v\n", err)
	}

	result.Success = true
	result.RetryCount = 0
	e.debugLog("ExecutePhase completed successfully - returning")
	return result, nil
}

// ExecuteWithRetry executes a command and automatically retries on failure
// This is a simplified version that doesn't require phase tracking
func (e *Executor) ExecuteWithRetry(command string, maxAttempts int) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := e.Claude.Execute(command)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < maxAttempts {
			fmt.Printf("Attempt %d/%d failed: %v\nRetrying...\n", attempt, maxAttempts, err)
		}
	}

	return fmt.Errorf("all %d attempts failed: %w", maxAttempts, lastErr)
}

// GetRetryState retrieves the current retry state for a spec/phase
func (e *Executor) GetRetryState(specName string, phase Phase) (*retry.RetryState, error) {
	return retry.LoadRetryState(e.StateDir, specName, string(phase), e.MaxRetries)
}

// ResetPhase resets the retry count for a specific phase
func (e *Executor) ResetPhase(specName string, phase Phase) error {
	return retry.ResetRetryCount(e.StateDir, specName, string(phase))
}

// ValidateSpec is a convenience wrapper for spec validation
func (e *Executor) ValidateSpec(specDir string) error {
	return validation.ValidateSpecFile(specDir)
}

// ValidatePlan is a convenience wrapper for plan validation
func (e *Executor) ValidatePlan(specDir string) error {
	return validation.ValidatePlanFile(specDir)
}

// ValidateTasks is a convenience wrapper for tasks validation
func (e *Executor) ValidateTasks(specDir string) error {
	return validation.ValidateTasksFile(specDir)
}

// ValidateTasksComplete checks if all tasks are completed
func (e *Executor) ValidateTasksComplete(tasksPath string) error {
	count, err := validation.CountUncheckedTasks(tasksPath)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("implementation incomplete: %d unchecked tasks remain", count)
	}

	return nil
}
