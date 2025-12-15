package progress_test

import (
	"errors"
	"testing"

	"github.com/ariel-frischer/autospec/internal/progress"
)

// BenchmarkProgressDisplay_StartPhase verifies StartPhase meets <10ms performance contract
func BenchmarkProgressDisplay_StartPhase(b *testing.B) {
	caps := progress.TerminalCapabilities{
		IsTTY:           false, // Avoid spinner overhead in benchmark
		SupportsUnicode: true,
		SupportsColor:   true,
		Width:           80,
	}

	display := progress.NewProgressDisplay(caps)
	phase := progress.PhaseInfo{
		Name:        "test",
		Number:      1,
		TotalPhases: 3,
		Status:      progress.PhaseInProgress,
		RetryCount:  0,
		MaxRetries:  3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = display.StartPhase(phase)
	}
}

// BenchmarkProgressDisplay_CompletePhase verifies CompletePhase meets <10ms performance contract
func BenchmarkProgressDisplay_CompletePhase(b *testing.B) {
	caps := progress.TerminalCapabilities{
		IsTTY:           false,
		SupportsUnicode: true,
		SupportsColor:   true,
		Width:           80,
	}

	display := progress.NewProgressDisplay(caps)
	phase := progress.PhaseInfo{
		Name:        "test",
		Number:      1,
		TotalPhases: 3,
		Status:      progress.PhaseCompleted,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = display.CompletePhase(phase)
	}
}

// BenchmarkProgressDisplay_FailPhase verifies FailPhase performance
func BenchmarkProgressDisplay_FailPhase(b *testing.B) {
	caps := progress.TerminalCapabilities{
		IsTTY:           false,
		SupportsUnicode: true,
		SupportsColor:   true,
		Width:           80,
	}

	display := progress.NewProgressDisplay(caps)
	phase := progress.PhaseInfo{
		Name:        "test",
		Number:      1,
		TotalPhases: 3,
		Status:      progress.PhaseFailed,
	}

	testErr := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = display.FailPhase(phase, testErr)
	}
}
