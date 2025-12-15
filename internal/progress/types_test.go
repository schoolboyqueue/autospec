package progress_test

import (
	"strings"
	"testing"

	"github.com/ariel-frischer/autospec/internal/progress"
)

// TestPhaseStatus_String tests the String() method of PhaseStatus enum
func TestPhaseStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status progress.PhaseStatus
		want   string
	}{
		{
			name:   "pending status",
			status: progress.PhasePending,
			want:   "pending",
		},
		{
			name:   "in_progress status",
			status: progress.PhaseInProgress,
			want:   "in_progress",
		},
		{
			name:   "completed status",
			status: progress.PhaseCompleted,
			want:   "completed",
		},
		{
			name:   "failed status",
			status: progress.PhaseFailed,
			want:   "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("PhaseStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPhaseInfo_Validate tests all validation rules for PhaseInfo
func TestPhaseInfo_Validate(t *testing.T) {
	tests := []struct {
		name    string
		phase   progress.PhaseInfo
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid phase info",
			phase: progress.PhaseInfo{
				Name:        "specify",
				Number:      1,
				TotalPhases: 3,
				Status:      progress.PhaseInProgress,
				RetryCount:  0,
				MaxRetries:  3,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			phase: progress.PhaseInfo{
				Name:        "",
				Number:      1,
				TotalPhases: 3,
				Status:      progress.PhaseInProgress,
			},
			wantErr: true,
			errMsg:  "phase name cannot be empty",
		},
		{
			name: "number less than or equal to zero",
			phase: progress.PhaseInfo{
				Name:        "test",
				Number:      0,
				TotalPhases: 3,
			},
			wantErr: true,
			errMsg:  "phase number must be > 0",
		},
		{
			name: "negative number",
			phase: progress.PhaseInfo{
				Name:        "test",
				Number:      -1,
				TotalPhases: 3,
			},
			wantErr: true,
			errMsg:  "phase number must be > 0",
		},
		{
			name: "number exceeds total phases",
			phase: progress.PhaseInfo{
				Name:        "test",
				Number:      4,
				TotalPhases: 3,
			},
			wantErr: true,
			errMsg:  "phase number cannot exceed total phases",
		},
		{
			name: "total phases less than or equal to zero",
			phase: progress.PhaseInfo{
				Name:        "test",
				Number:      1,
				TotalPhases: 0,
			},
			wantErr: true,
			errMsg:  "total phases must be > 0",
		},
		{
			name: "negative total phases",
			phase: progress.PhaseInfo{
				Name:        "test",
				Number:      1,
				TotalPhases: -1,
			},
			wantErr: true,
			errMsg:  "total phases must be > 0",
		},
		{
			name: "negative retry count",
			phase: progress.PhaseInfo{
				Name:        "test",
				Number:      1,
				TotalPhases: 3,
				RetryCount:  -1,
			},
			wantErr: true,
			errMsg:  "retry count cannot be negative",
		},
		{
			name: "negative max retries",
			phase: progress.PhaseInfo{
				Name:        "test",
				Number:      1,
				TotalPhases: 3,
				RetryCount:  0,
				MaxRetries:  -1,
			},
			wantErr: true,
			errMsg:  "max retries cannot be negative",
		},
		{
			name: "valid with retry attempt",
			phase: progress.PhaseInfo{
				Name:        "plan",
				Number:      2,
				TotalPhases: 4,
				Status:      progress.PhaseInProgress,
				RetryCount:  1,
				MaxRetries:  3,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.phase.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("PhaseInfo.Validate() error = nil, want error containing %q", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("PhaseInfo.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("PhaseInfo.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}
