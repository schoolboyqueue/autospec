package cli

import (
	"testing"

	"github.com/ariel-frischer/autospec/internal/workflow"
)

func TestStageConfigFromFlags(t *testing.T) {
	tests := []struct {
		name     string
		config   *workflow.StageConfig
		expected []workflow.Stage
	}{
		{
			name: "core stages only (-spti)",
			config: &workflow.StageConfig{
				Specify:   true,
				Plan:      true,
				Tasks:     true,
				Implement: true,
			},
			expected: []workflow.Stage{
				workflow.StageSpecify,
				workflow.StagePlan,
				workflow.StageTasks,
				workflow.StageImplement,
			},
		},
		{
			name: "constitution and specify (-ns)",
			config: &workflow.StageConfig{
				Constitution: true,
				Specify:      true,
			},
			expected: []workflow.Stage{
				workflow.StageConstitution,
				workflow.StageSpecify,
			},
		},
		{
			name: "specify, clarify, plan (-srp)",
			config: &workflow.StageConfig{
				Specify: true,
				Clarify: true,
				Plan:    true,
			},
			expected: []workflow.Stage{
				workflow.StageSpecify,
				workflow.StageClarify,
				workflow.StagePlan,
			},
		},
		{
			name: "tasks, checklist, analyze, implement (-tlzi)",
			config: &workflow.StageConfig{
				Tasks:     true,
				Checklist: true,
				Analyze:   true,
				Implement: true,
			},
			expected: []workflow.Stage{
				workflow.StageTasks,
				workflow.StageChecklist,
				workflow.StageAnalyze,
				workflow.StageImplement,
			},
		},
		{
			name: "all stages with checklist (-a -l) - core + optional",
			config: &workflow.StageConfig{
				Specify:   true,
				Plan:      true,
				Tasks:     true,
				Implement: true,
				Checklist: true,
			},
			expected: []workflow.Stage{
				workflow.StageSpecify,
				workflow.StagePlan,
				workflow.StageTasks,
				workflow.StageChecklist,
				workflow.StageImplement,
			},
		},
		{
			name: "all 8 stages in canonical order",
			config: &workflow.StageConfig{
				Constitution: true,
				Specify:      true,
				Clarify:      true,
				Plan:         true,
				Tasks:        true,
				Checklist:    true,
				Analyze:      true,
				Implement:    true,
			},
			expected: []workflow.Stage{
				workflow.StageConstitution,
				workflow.StageSpecify,
				workflow.StageClarify,
				workflow.StagePlan,
				workflow.StageTasks,
				workflow.StageChecklist,
				workflow.StageAnalyze,
				workflow.StageImplement,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetCanonicalOrder()
			if len(got) != len(tt.expected) {
				t.Errorf("GetCanonicalOrder() returned %d stages, want %d", len(got), len(tt.expected))
				return
			}
			for i, stage := range got {
				if stage != tt.expected[i] {
					t.Errorf("GetCanonicalOrder()[%d] = %v, want %v", i, stage, tt.expected[i])
				}
			}
		})
	}
}

func TestOptionalStagesWithAll(t *testing.T) {
	// Test that optional stages can be combined with -a flag
	// -a sets core stages, optional stages are added separately
	config := &workflow.StageConfig{}
	config.SetAll() // Sets core stages only

	// Add optional stages
	config.Checklist = true
	config.Analyze = true

	stages := config.GetCanonicalOrder()

	// Should be: specify, plan, tasks, checklist, analyze, implement
	expected := []workflow.Stage{
		workflow.StageSpecify,
		workflow.StagePlan,
		workflow.StageTasks,
		workflow.StageChecklist,
		workflow.StageAnalyze,
		workflow.StageImplement,
	}

	if len(stages) != len(expected) {
		t.Errorf("Expected %d stages, got %d", len(expected), len(stages))
		return
	}

	for i, stage := range stages {
		if stage != expected[i] {
			t.Errorf("Stage %d: expected %s, got %s", i, expected[i], stage)
		}
	}
}

func TestOptionalStageHasAnyStage(t *testing.T) {
	// Test that HasAnyStage returns true for optional stages only
	tests := []struct {
		name     string
		config   *workflow.StageConfig
		expected bool
	}{
		{
			name:     "no stages",
			config:   &workflow.StageConfig{},
			expected: false,
		},
		{
			name:     "only constitution",
			config:   &workflow.StageConfig{Constitution: true},
			expected: true,
		},
		{
			name:     "only clarify",
			config:   &workflow.StageConfig{Clarify: true},
			expected: true,
		},
		{
			name:     "only checklist",
			config:   &workflow.StageConfig{Checklist: true},
			expected: true,
		},
		{
			name:     "only analyze",
			config:   &workflow.StageConfig{Analyze: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.HasAnyStage(); got != tt.expected {
				t.Errorf("HasAnyStage() = %v, want %v", got, tt.expected)
			}
		})
	}
}
