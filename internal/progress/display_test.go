package progress_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ariel-frischer/autospec/internal/progress"
)

// captureOutput captures stdout during function execution
func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestProgressDisplay_StartPhase tests phase counter rendering
func TestProgressDisplay_StartPhase(t *testing.T) {
	tests := []struct {
		name         string
		capabilities progress.TerminalCapabilities
		phase        progress.PhaseInfo
		wantContains []string
		wantErr      bool
	}{
		{
			name: "TTY mode with Unicode - first phase",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           true,
				SupportsUnicode: true,
				SupportsColor:   true,
				Width:           80,
			},
			phase: progress.PhaseInfo{
				Name:        "specify",
				Number:      1,
				TotalPhases: 3,
				Status:      progress.PhaseInProgress,
				RetryCount:  0,
				MaxRetries:  3,
			},
			wantContains: []string{"[1/3]", "specify"},
			wantErr:      false,
		},
		{
			name: "non-TTY mode - second phase",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           false,
				SupportsUnicode: false,
				SupportsColor:   false,
				Width:           0,
			},
			phase: progress.PhaseInfo{
				Name:        "plan",
				Number:      2,
				TotalPhases: 3,
				Status:      progress.PhaseInProgress,
				RetryCount:  0,
				MaxRetries:  3,
			},
			wantContains: []string{"[2/3]", "Plan"},
			wantErr:      false,
		},
		{
			name: "retry attempt - show retry count",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           true,
				SupportsUnicode: true,
				SupportsColor:   true,
				Width:           80,
			},
			phase: progress.PhaseInfo{
				Name:        "tasks",
				Number:      3,
				TotalPhases: 3,
				Status:      progress.PhaseInProgress,
				RetryCount:  1,
				MaxRetries:  3,
			},
			wantContains: []string{"[3/3]", "tasks", "(retry 2/3)"},
			wantErr:      false,
		},
		{
			name: "invalid phase - empty name",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           true,
				SupportsUnicode: true,
				SupportsColor:   true,
				Width:           80,
			},
			phase: progress.PhaseInfo{
				Name:        "",
				Number:      1,
				TotalPhases: 3,
			},
			wantErr: true,
		},
		{
			name: "four-phase workflow - implement phase",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           true,
				SupportsUnicode: true,
				SupportsColor:   true,
				Width:           80,
			},
			phase: progress.PhaseInfo{
				Name:        "implement",
				Number:      4,
				TotalPhases: 4,
				Status:      progress.PhaseInProgress,
				RetryCount:  0,
				MaxRetries:  3,
			},
			wantContains: []string{"[4/4]", "implement"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display := progress.NewProgressDisplay(tt.capabilities)

			var output string
			var err error

			if tt.capabilities.IsTTY {
				// For TTY mode, spinner starts, so we just check error
				err = display.StartPhase(tt.phase)
			} else {
				// For non-TTY mode, capture stdout
				output = captureOutput(func() {
					err = display.StartPhase(tt.phase)
				})
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("StartPhase() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Errorf("StartPhase() unexpected error = %v", err)
				return
			}

			// For non-TTY mode, verify output contains expected strings
			if !tt.capabilities.IsTTY {
				for _, want := range tt.wantContains {
					if !strings.Contains(output, want) {
						t.Errorf("StartPhase() output = %q, want to contain %q", output, want)
					}
				}
			}
		})
	}
}

// TestProgressDisplay_UpdateRetry tests retry count display
func TestProgressDisplay_UpdateRetry(t *testing.T) {
	caps := progress.TerminalCapabilities{
		IsTTY:           false,
		SupportsUnicode: false,
		SupportsColor:   false,
	}

	phase := progress.PhaseInfo{
		Name:        "specify",
		Number:      1,
		TotalPhases: 3,
		Status:      progress.PhaseInProgress,
		RetryCount:  2,
		MaxRetries:  3,
	}

	display := progress.NewProgressDisplay(caps)

	output := captureOutput(func() {
		_ = display.UpdateRetry(phase)
	})

	// Should show retry 3/3 (RetryCount is 0-indexed, display is 1-indexed)
	if !strings.Contains(output, "(retry 3/3)") {
		t.Errorf("UpdateRetry() output = %q, want to contain '(retry 3/3)'", output)
	}
}

// TestProgressDisplay_CompletePhase tests completion checkmarks (User Story 3)
func TestProgressDisplay_CompletePhase(t *testing.T) {
	tests := []struct {
		name         string
		capabilities progress.TerminalCapabilities
		phase        progress.PhaseInfo
		wantContains []string
	}{
		{
			name: "Unicode checkmark with color",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           true,
				SupportsUnicode: true,
				SupportsColor:   true,
				Width:           80,
			},
			phase: progress.PhaseInfo{
				Name:        "specify",
				Number:      1,
				TotalPhases: 3,
				Status:      progress.PhaseCompleted,
			},
			wantContains: []string{"✓", "[1/3]", "Specify", "complete"},
		},
		{
			name: "ASCII checkmark without color",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           true,
				SupportsUnicode: false,
				SupportsColor:   false,
				Width:           80,
			},
			phase: progress.PhaseInfo{
				Name:        "plan",
				Number:      2,
				TotalPhases: 3,
				Status:      progress.PhaseCompleted,
			},
			wantContains: []string{"[OK]", "[2/3]", "Plan", "complete"},
		},
		{
			name: "non-TTY mode completion",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           false,
				SupportsUnicode: false,
				SupportsColor:   false,
			},
			phase: progress.PhaseInfo{
				Name:        "tasks",
				Number:      3,
				TotalPhases: 3,
				Status:      progress.PhaseCompleted,
			},
			wantContains: []string{"[OK]", "[3/3]", "Tasks", "complete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display := progress.NewProgressDisplay(tt.capabilities)

			output := captureOutput(func() {
				_ = display.CompletePhase(tt.phase)
			})

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("CompletePhase() output = %q, want to contain %q", output, want)
				}
			}
		})
	}
}

// TestProgressDisplay_FailPhase tests failure indicators (User Story 3)
func TestProgressDisplay_FailPhase(t *testing.T) {
	tests := []struct {
		name         string
		capabilities progress.TerminalCapabilities
		phase        progress.PhaseInfo
		err          error
		wantContains []string
	}{
		{
			name: "Unicode failure mark with color",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           true,
				SupportsUnicode: true,
				SupportsColor:   true,
				Width:           80,
			},
			phase: progress.PhaseInfo{
				Name:        "specify",
				Number:      1,
				TotalPhases: 3,
				Status:      progress.PhaseFailed,
			},
			err:          fmt.Errorf("validation failed"),
			wantContains: []string{"✗", "[1/3]", "Specify", "failed", "validation failed"},
		},
		{
			name: "ASCII failure mark without color",
			capabilities: progress.TerminalCapabilities{
				IsTTY:           true,
				SupportsUnicode: false,
				SupportsColor:   false,
				Width:           80,
			},
			phase: progress.PhaseInfo{
				Name:        "plan",
				Number:      2,
				TotalPhases: 3,
				Status:      progress.PhaseFailed,
			},
			err:          fmt.Errorf("file not found"),
			wantContains: []string{"[FAIL]", "[2/3]", "Plan", "failed", "file not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display := progress.NewProgressDisplay(tt.capabilities)

			output := captureOutput(func() {
				_ = display.FailPhase(tt.phase, tt.err)
			})

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("FailPhase() output = %q, want to contain %q", output, want)
				}
			}
		})
	}
}

// TestSpinnerLifecycle tests spinner start/stop behavior (User Story 2)
func TestSpinnerLifecycle(t *testing.T) {
	// TTY mode - spinner should start
	capsTTY := progress.TerminalCapabilities{
		IsTTY:           true,
		SupportsUnicode: true,
		SupportsColor:   true,
		Width:           80,
	}

	phase := progress.PhaseInfo{
		Name:        "specify",
		Number:      1,
		TotalPhases: 3,
		Status:      progress.PhaseInProgress,
	}

	display := progress.NewProgressDisplay(capsTTY)

	// Start phase - spinner starts
	err := display.StartPhase(phase)
	if err != nil {
		t.Fatalf("StartPhase() unexpected error = %v", err)
	}

	// Complete phase - spinner should stop
	output := captureOutput(func() {
		_ = display.CompletePhase(phase)
	})

	if !strings.Contains(output, "✓") {
		t.Errorf("CompletePhase() output = %q, want to contain checkmark", output)
	}
}

// TestSpinnerDisabledNonTTY tests spinner is disabled in non-TTY mode (User Story 2)
func TestSpinnerDisabledNonTTY(t *testing.T) {
	capsNonTTY := progress.TerminalCapabilities{
		IsTTY:           false,
		SupportsUnicode: false,
		SupportsColor:   false,
	}

	phase := progress.PhaseInfo{
		Name:        "plan",
		Number:      1,
		TotalPhases: 3,
		Status:      progress.PhaseInProgress,
	}

	display := progress.NewProgressDisplay(capsNonTTY)

	output := captureOutput(func() {
		_ = display.StartPhase(phase)
	})

	// Non-TTY mode should just print the message, no spinner
	if !strings.Contains(output, "[1/3]") || !strings.Contains(output, "Plan") {
		t.Errorf("StartPhase() non-TTY output = %q, want phase message", output)
	}
}
