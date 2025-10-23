package progress

import "errors"

// PhaseStatus represents the execution state of a workflow phase
type PhaseStatus int

const (
	// PhasePending indicates the phase has not started yet
	PhasePending PhaseStatus = iota
	// PhaseInProgress indicates the phase is currently running
	PhaseInProgress
	// PhaseCompleted indicates the phase finished successfully
	PhaseCompleted
	// PhaseFailed indicates the phase failed with an error
	PhaseFailed
)

// String returns the string representation of PhaseStatus
func (s PhaseStatus) String() string {
	switch s {
	case PhasePending:
		return "pending"
	case PhaseInProgress:
		return "in_progress"
	case PhaseCompleted:
		return "completed"
	case PhaseFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// PhaseInfo represents metadata about a workflow phase for progress display
type PhaseInfo struct {
	// Name is the human-readable phase name (e.g., "specify", "plan", "tasks", "implement")
	Name string
	// Number is the current phase number (1-based index)
	Number int
	// TotalPhases is the total number of phases in the workflow
	TotalPhases int
	// Status is the current execution status
	Status PhaseStatus
	// RetryCount is the number of retry attempts (0 if first attempt)
	RetryCount int
	// MaxRetries is the maximum retry attempts allowed
	MaxRetries int
}

// Validate checks that all PhaseInfo fields meet validation requirements
func (p PhaseInfo) Validate() error {
	if p.Name == "" {
		return errors.New("phase name cannot be empty")
	}
	if p.Number <= 0 {
		return errors.New("phase number must be > 0")
	}
	if p.TotalPhases <= 0 {
		return errors.New("total phases must be > 0")
	}
	if p.Number > p.TotalPhases {
		return errors.New("phase number cannot exceed total phases")
	}
	if p.RetryCount < 0 {
		return errors.New("retry count cannot be negative")
	}
	if p.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}
	return nil
}

// TerminalCapabilities encapsulates detected terminal features
type TerminalCapabilities struct {
	// IsTTY indicates whether stdout is a terminal (vs pipe/redirect)
	IsTTY bool
	// SupportsColor indicates whether terminal supports ANSI color codes
	SupportsColor bool
	// SupportsUnicode indicates whether terminal supports Unicode characters
	SupportsUnicode bool
	// Width is the terminal width in columns (0 if unknown/pipe)
	Width int
}

// ProgressSymbols defines the character set for visual indicators
type ProgressSymbols struct {
	// Checkmark is the success indicator ("✓" or "[OK]")
	Checkmark string
	// Failure is the failure indicator ("✗" or "[FAIL]")
	Failure string
	// SpinnerSet is the index into spinner.CharSets
	SpinnerSet int
}
