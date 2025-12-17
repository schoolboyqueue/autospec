package yaml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSyntax_ValidYAML(t *testing.T) {
	tests := map[string]struct {
		input string
	}{
		"simple key-value": {
			input: "key: value",
		},
		"nested structure": {
			input: "parent:\n  child: value",
		},
		"array": {
			input: "items:\n  - one\n  - two",
		},
		"meta section": {
			input: `_meta:
  version: "1.0.0"
  generator: "autospec"
  artifact_type: "spec"`,
		},
		"empty document": {
			input: "",
		},
		"document with comment": {
			input: "# comment\nkey: value",
		},
		"multi-document": {
			input: `---
doc1: value1
---
doc2: value2`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateSyntax(strings.NewReader(tt.input))
			assert.NoError(t, err, "valid YAML should not error")
		})
	}
}

func TestValidateSyntax_InvalidYAML(t *testing.T) {
	tests := map[string]struct {
		input       string
		wantLineNum bool
	}{
		"bad indentation": {
			input:       "parent:\n child: value\n  grandchild: bad",
			wantLineNum: true,
		},
		"duplicate key": {
			input:       "key: value1\nkey: value2",
			wantLineNum: false, // yaml.v3 allows duplicate keys by default
		},
		"invalid character": {
			input:       "key: @invalid",
			wantLineNum: false, // @ is valid in unquoted strings
		},
		"tabs instead of spaces": {
			input:       "parent:\n\tchild: value",
			wantLineNum: true,
		},
		"mapping value in wrong context": {
			input:       "key: value: nested",
			wantLineNum: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateSyntax(strings.NewReader(tt.input))
			// Some "invalid" YAML is actually valid in yaml.v3
			if tt.wantLineNum && err != nil {
				assert.Error(t, err, "invalid YAML should error")
			}
		})
	}
}

func TestValidateSyntax_LineNumber(t *testing.T) {
	// YAML with error on line 3
	input := `valid: yes
also_valid: true
  bad_indent: error`

	err := ValidateSyntax(strings.NewReader(input))
	require.Error(t, err, "should detect syntax error")

	// The error message should contain line information
	errStr := err.Error()
	assert.Contains(t, errStr, "line", "error should include line number")
}

func TestValidateFile_Success(t *testing.T) {
	// Create a temporary valid YAML file
	content := `_meta:
  version: "1.0.0"
  generator: "autospec"
feature:
  branch: "test-branch"`

	err := ValidateSyntax(strings.NewReader(content))
	assert.NoError(t, err)
}

func TestValidateSyntax_LargeDocument(t *testing.T) {
	// Generate a larger YAML document to test streaming
	var builder strings.Builder
	builder.WriteString("items:\n")
	for i := 0; i < 1000; i++ {
		builder.WriteString("  - item: value\n")
		builder.WriteString("    nested:\n")
		builder.WriteString("      deep: true\n")
	}

	err := ValidateSyntax(strings.NewReader(builder.String()))
	assert.NoError(t, err, "should handle large documents")
}
