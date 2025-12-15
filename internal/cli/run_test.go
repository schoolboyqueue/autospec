package cli

import (
	"testing"

	"github.com/ariel-frischer/autospec/internal/workflow"
)

func TestPhaseConfigFromFlags(t *testing.T) {
	tests := []struct {
		name     string
		config   *workflow.PhaseConfig
		expected []workflow.Phase
	}{
		{
			name: "core phases only (-spti)",
			config: &workflow.PhaseConfig{
				Specify:   true,
				Plan:      true,
				Tasks:     true,
				Implement: true,
			},
			expected: []workflow.Phase{
				workflow.PhaseSpecify,
				workflow.PhasePlan,
				workflow.PhaseTasks,
				workflow.PhaseImplement,
			},
		},
		{
			name: "constitution and specify (-ns)",
			config: &workflow.PhaseConfig{
				Constitution: true,
				Specify:      true,
			},
			expected: []workflow.Phase{
				workflow.PhaseConstitution,
				workflow.PhaseSpecify,
			},
		},
		{
			name: "specify, clarify, plan (-srp)",
			config: &workflow.PhaseConfig{
				Specify: true,
				Clarify: true,
				Plan:    true,
			},
			expected: []workflow.Phase{
				workflow.PhaseSpecify,
				workflow.PhaseClarify,
				workflow.PhasePlan,
			},
		},
		{
			name: "tasks, checklist, analyze, implement (-tlzi)",
			config: &workflow.PhaseConfig{
				Tasks:     true,
				Checklist: true,
				Analyze:   true,
				Implement: true,
			},
			expected: []workflow.Phase{
				workflow.PhaseTasks,
				workflow.PhaseChecklist,
				workflow.PhaseAnalyze,
				workflow.PhaseImplement,
			},
		},
		{
			name: "all phases with checklist (-a -l) - core + optional",
			config: &workflow.PhaseConfig{
				Specify:   true,
				Plan:      true,
				Tasks:     true,
				Implement: true,
				Checklist: true,
			},
			expected: []workflow.Phase{
				workflow.PhaseSpecify,
				workflow.PhasePlan,
				workflow.PhaseTasks,
				workflow.PhaseChecklist,
				workflow.PhaseImplement,
			},
		},
		{
			name: "all 8 phases in canonical order",
			config: &workflow.PhaseConfig{
				Constitution: true,
				Specify:      true,
				Clarify:      true,
				Plan:         true,
				Tasks:        true,
				Checklist:    true,
				Analyze:      true,
				Implement:    true,
			},
			expected: []workflow.Phase{
				workflow.PhaseConstitution,
				workflow.PhaseSpecify,
				workflow.PhaseClarify,
				workflow.PhasePlan,
				workflow.PhaseTasks,
				workflow.PhaseChecklist,
				workflow.PhaseAnalyze,
				workflow.PhaseImplement,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetCanonicalOrder()
			if len(got) != len(tt.expected) {
				t.Errorf("GetCanonicalOrder() returned %d phases, want %d", len(got), len(tt.expected))
				return
			}
			for i, phase := range got {
				if phase != tt.expected[i] {
					t.Errorf("GetCanonicalOrder()[%d] = %v, want %v", i, phase, tt.expected[i])
				}
			}
		})
	}
}

func TestOptionalPhasesWithAll(t *testing.T) {
	// Test that optional phases can be combined with -a flag
	// -a sets core phases, optional phases are added separately
	config := &workflow.PhaseConfig{}
	config.SetAll() // Sets core phases only

	// Add optional phases
	config.Checklist = true
	config.Analyze = true

	phases := config.GetCanonicalOrder()

	// Should be: specify, plan, tasks, checklist, analyze, implement
	expected := []workflow.Phase{
		workflow.PhaseSpecify,
		workflow.PhasePlan,
		workflow.PhaseTasks,
		workflow.PhaseChecklist,
		workflow.PhaseAnalyze,
		workflow.PhaseImplement,
	}

	if len(phases) != len(expected) {
		t.Errorf("Expected %d phases, got %d", len(expected), len(phases))
		return
	}

	for i, phase := range phases {
		if phase != expected[i] {
			t.Errorf("Phase %d: expected %s, got %s", i, expected[i], phase)
		}
	}
}

func TestOptionalPhaseHasAnyPhase(t *testing.T) {
	// Test that HasAnyPhase returns true for optional phases only
	tests := []struct {
		name     string
		config   *workflow.PhaseConfig
		expected bool
	}{
		{
			name:     "no phases",
			config:   &workflow.PhaseConfig{},
			expected: false,
		},
		{
			name:     "only constitution",
			config:   &workflow.PhaseConfig{Constitution: true},
			expected: true,
		},
		{
			name:     "only clarify",
			config:   &workflow.PhaseConfig{Clarify: true},
			expected: true,
		},
		{
			name:     "only checklist",
			config:   &workflow.PhaseConfig{Checklist: true},
			expected: true,
		},
		{
			name:     "only analyze",
			config:   &workflow.PhaseConfig{Analyze: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.HasAnyPhase(); got != tt.expected {
				t.Errorf("HasAnyPhase() = %v, want %v", got, tt.expected)
			}
		})
	}
}
