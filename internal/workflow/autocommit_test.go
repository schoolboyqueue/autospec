package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectableInstruction(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		inst     InjectableInstruction
		wantName string
		wantHint string
	}{
		"all fields populated": {
			inst: InjectableInstruction{
				Name:        "AutoCommit",
				DisplayHint: "post-work git commit with conventional format",
				Content:     "Do the commit",
			},
			wantName: "AutoCommit",
			wantHint: "post-work git commit with conventional format",
		},
		"name only with content": {
			inst: InjectableInstruction{
				Name:    "TestInstruction",
				Content: "Test content",
			},
			wantName: "TestInstruction",
			wantHint: "",
		},
		"empty display hint is valid": {
			inst: InjectableInstruction{
				Name:        "MinimalInst",
				DisplayHint: "",
				Content:     "Content here",
			},
			wantName: "MinimalInst",
			wantHint: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantName, tc.inst.Name)
			assert.Equal(t, tc.wantHint, tc.inst.DisplayHint)
			assert.NotEmpty(t, tc.inst.Content)
		})
	}
}

func TestInjectInstructions(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		command      string
		instructions []InjectableInstruction
		wantContains []string
		wantExact    string
	}{
		"empty instruction list returns original": {
			command:      "original command",
			instructions: []InjectableInstruction{},
			wantExact:    "original command",
		},
		"nil instruction list returns original": {
			command:      "original command",
			instructions: nil,
			wantExact:    "original command",
		},
		"single instruction with markers": {
			command: "base command",
			instructions: []InjectableInstruction{
				{Name: "AutoCommit", Content: "Commit instructions"},
			},
			wantContains: []string{
				"base command",
				"<!-- AUTOSPEC_INJECT:AutoCommit -->",
				"Commit instructions",
				"<!-- /AUTOSPEC_INJECT:AutoCommit -->",
			},
		},
		"multiple instructions": {
			command: "multi command",
			instructions: []InjectableInstruction{
				{Name: "First", Content: "First content"},
				{Name: "Second", Content: "Second content"},
			},
			wantContains: []string{
				"multi command",
				"<!-- AUTOSPEC_INJECT:First -->",
				"First content",
				"<!-- /AUTOSPEC_INJECT:First -->",
				"<!-- AUTOSPEC_INJECT:Second -->",
				"Second content",
				"<!-- /AUTOSPEC_INJECT:Second -->",
			},
		},
		"skips instruction with empty name": {
			command: "cmd",
			instructions: []InjectableInstruction{
				{Name: "", Content: "Should be skipped"},
				{Name: "Valid", Content: "Should be included"},
			},
			wantContains: []string{
				"<!-- AUTOSPEC_INJECT:Valid -->",
				"Should be included",
			},
		},
		"skips instruction with empty content": {
			command: "cmd",
			instructions: []InjectableInstruction{
				{Name: "EmptyContent", Content: ""},
				{Name: "Valid", Content: "Valid content"},
			},
			wantContains: []string{
				"<!-- AUTOSPEC_INJECT:Valid -->",
				"Valid content",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := InjectInstructions(tc.command, tc.instructions)

			if tc.wantExact != "" {
				assert.Equal(t, tc.wantExact, result)
				return
			}

			for _, want := range tc.wantContains {
				assert.Contains(t, result, want,
					"result should contain: %s", want)
			}
		})
	}
}

func TestInjectInstructionsMarkerFormat(t *testing.T) {
	t.Parallel()

	command := "test"
	instructions := []InjectableInstruction{
		{Name: "TestMarker", Content: "Test content"},
	}

	result := InjectInstructions(command, instructions)

	// Verify marker structure is correct for parsing
	require.Contains(t, result, "<!-- AUTOSPEC_INJECT:TestMarker -->")
	require.Contains(t, result, "<!-- /AUTOSPEC_INJECT:TestMarker -->")

	// Verify content is between markers
	startIdx := strings.Index(result, "<!-- AUTOSPEC_INJECT:TestMarker -->")
	endIdx := strings.Index(result, "<!-- /AUTOSPEC_INJECT:TestMarker -->")
	contentIdx := strings.Index(result, "Test content")

	assert.Less(t, startIdx, contentIdx, "start marker should come before content")
	assert.Less(t, contentIdx, endIdx, "content should come before end marker")
}

func TestInjectInstructionsPreservesOriginal(t *testing.T) {
	t.Parallel()

	// Original command with complex content
	original := `Some complex command
with multiple lines
and special chars: $VAR && || > <`

	instructions := []InjectableInstruction{
		{Name: "Append", Content: "Appended content"},
	}

	result := InjectInstructions(original, instructions)

	// Original should be fully preserved at the start
	assert.True(t, strings.HasPrefix(result, original),
		"result should start with original command")
}

func TestBuildAutoCommitInstructions(t *testing.T) {
	t.Parallel()

	instructions := BuildAutoCommitInstructions()

	// Verify instructions are non-empty
	assert.NotEmpty(t, instructions, "instructions should not be empty")

	t.Run("contains gitignore guidance", func(t *testing.T) {
		t.Parallel()

		// Should contain .gitignore section
		assert.Contains(t, instructions, ".gitignore",
			"instructions should mention .gitignore")
		assert.Contains(t, instructions, "Update .gitignore",
			"instructions should include .gitignore update section")

		// Should include common ignorable patterns
		patterns := []string{
			"node_modules",
			"__pycache__",
			".venv",
			"dist/",
			"build/",
			".DS_Store",
			".env",
		}
		for _, pattern := range patterns {
			assert.Contains(t, instructions, pattern,
				"instructions should include common pattern: %s", pattern)
		}
	})

	t.Run("contains staging rules", func(t *testing.T) {
		t.Parallel()

		// Should contain staging guidance
		assert.Contains(t, instructions, "Stage",
			"instructions should mention staging")
		assert.Contains(t, instructions, "git add",
			"instructions should include git add command")
		assert.Contains(t, instructions, "git status",
			"instructions should include git status for verification")

		// Should mention what NOT to stage
		assert.Contains(t, instructions, "Do NOT stage",
			"instructions should mention files not to stage")
	})

	t.Run("contains conventional commit format", func(t *testing.T) {
		t.Parallel()

		// Should contain conventional commit section
		assert.Contains(t, instructions, "Conventional Commit",
			"instructions should mention conventional commit")
		assert.Contains(t, instructions, "type(scope): description",
			"instructions should include commit format template")

		// Should include common commit types
		commitTypes := []string{
			"feat:",
			"fix:",
			"docs:",
			"style:",
			"refactor:",
			"test:",
			"chore:",
		}
		for _, ct := range commitTypes {
			assert.Contains(t, instructions, ct,
				"instructions should include commit type: %s", ct)
		}

		// Should include scope guidance
		assert.Contains(t, instructions, "scope",
			"instructions should mention scope")
		assert.Contains(t, instructions, "git commit",
			"instructions should include git commit command")
	})

	t.Run("is agent-agnostic", func(t *testing.T) {
		t.Parallel()

		// Should NOT contain agent-specific references
		agentSpecificTerms := []string{
			"Claude",
			"Gemini",
			"GPT",
			"OpenAI",
			"Anthropic",
			"Google AI",
			"Copilot",
		}
		instructionsLower := strings.ToLower(instructions)
		for _, term := range agentSpecificTerms {
			assert.NotContains(t, instructionsLower, strings.ToLower(term),
				"instructions should not contain agent-specific term: %s", term)
		}
	})

	t.Run("handles edge cases", func(t *testing.T) {
		t.Parallel()

		// Should mention what to do when there are no changes
		assert.Contains(t, instructions, "no changes",
			"instructions should handle no-changes case")

		// Should mention detached HEAD state
		assert.Contains(t, instructions, "detached HEAD",
			"instructions should handle detached HEAD case")
	})
}

func TestBuildAutoCommitInstructionsIdempotent(t *testing.T) {
	t.Parallel()

	// Calling the function multiple times should return the same result
	first := BuildAutoCommitInstructions()
	second := BuildAutoCommitInstructions()

	assert.Equal(t, first, second,
		"BuildAutoCommitInstructions should be idempotent")
}

func TestAutoCommitInstructionsStructure(t *testing.T) {
	t.Parallel()

	instructions := BuildAutoCommitInstructions()

	tests := map[string]struct {
		section     string
		description string
	}{
		"has step 1 header": {
			section:     "Step 1",
			description: "gitignore update step",
		},
		"has step 2 header": {
			section:     "Step 2",
			description: "staging step",
		},
		"has step 3 header": {
			section:     "Step 3",
			description: "commit creation step",
		},
		"has important notes": {
			section:     "Important Notes",
			description: "edge case handling notes",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, instructions, tc.section,
				"instructions should contain %s for %s", tc.section, tc.description)
		})
	}
}
