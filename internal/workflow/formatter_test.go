package workflow

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ariel-frischer/autospec/internal/config"
	"github.com/ariel-frischer/claude-clean/display"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStreamFormatter(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		style config.OutputStyle
	}{
		"default style": {style: config.OutputStyleDefault},
		"compact style": {style: config.OutputStyleCompact},
		"minimal style": {style: config.OutputStyleMinimal},
		"plain style":   {style: config.OutputStylePlain},
		"raw style":     {style: config.OutputStyleRaw},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			f := NewStreamFormatter(tt.style, buf)

			assert.NotNil(t, f)
			assert.Equal(t, tt.style, f.style)
			assert.Equal(t, 0, f.lineNum)
			assert.NotNil(t, f.config)
		})
	}
}

func TestMapStyleToConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		style    config.OutputStyle
		wantType display.OutputStyle
	}{
		"default maps to StyleDefault": {
			style:    config.OutputStyleDefault,
			wantType: display.StyleDefault,
		},
		"compact maps to StyleCompact": {
			style:    config.OutputStyleCompact,
			wantType: display.StyleCompact,
		},
		"minimal maps to StyleMinimal": {
			style:    config.OutputStyleMinimal,
			wantType: display.StyleMinimal,
		},
		"plain maps to StylePlain": {
			style:    config.OutputStylePlain,
			wantType: display.StylePlain,
		},
		"raw maps to StyleDefault as fallback": {
			style:    config.OutputStyleRaw,
			wantType: display.StyleDefault,
		},
		"empty string maps to StyleDefault": {
			style:    "",
			wantType: display.StyleDefault,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := mapStyleToConfig(tt.style)

			assert.NotNil(t, cfg)
			assert.Equal(t, tt.wantType, cfg.Style)
			assert.False(t, cfg.Verbose)
			assert.False(t, cfg.ShowLineNum)
		})
	}
}

func TestStreamFormatter_ProcessLine_EmptyLines(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
	}{
		"empty string":     {input: ""},
		"whitespace only":  {input: "   "},
		"tabs only":        {input: "\t\t"},
		"mixed whitespace": {input: " \t \n "},
		"newlines only":    {input: "\n\n"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			f := NewStreamFormatter(config.OutputStyleDefault, buf)

			f.ProcessLine(tt.input)

			// Empty lines should not produce output
			assert.Empty(t, buf.String())
			assert.Equal(t, 0, f.lineNum)
		})
	}
}

func TestStreamFormatter_ProcessLine_RawStyle(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input       string
		wantContain string
	}{
		"valid JSON passed through": {
			input:       `{"type":"message"}`,
			wantContain: `{"type":"message"}`,
		},
		"invalid JSON passed through": {
			input:       "not json at all",
			wantContain: "not json at all",
		},
		"line with system reminder passed through": {
			input:       `{"message":"<system-reminder>keep this</system-reminder>"}`,
			wantContain: `<system-reminder>`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			f := NewStreamFormatter(config.OutputStyleRaw, buf)

			f.ProcessLine(tt.input)

			output := buf.String()
			assert.Contains(t, output, tt.wantContain)
			assert.True(t, strings.HasSuffix(output, "\n"))
		})
	}
}

func TestStreamFormatter_ProcessLine_InvalidJSON(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input       string
		wantContain string
	}{
		"plain text": {
			input:       "Just some plain text",
			wantContain: "Just some plain text",
		},
		"truncated JSON": {
			input:       `{"type": "message`,
			wantContain: `{"type": "message`,
		},
		"partial JSON array": {
			input:       `[1, 2, 3`,
			wantContain: `[1, 2, 3`,
		},
		"non-UTF8 fallback": {
			input:       "text with\xffbad byte",
			wantContain: "text with",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			f := NewStreamFormatter(config.OutputStyleDefault, buf)

			f.ProcessLine(tt.input)

			output := buf.String()
			// Invalid JSON should fall back to raw output
			assert.Contains(t, output, tt.wantContain)
		})
	}
}

func TestStreamFormatter_ProcessLine_ValidJSON(t *testing.T) {
	t.Parallel()

	// Sample JSONL from Claude's stream-json output
	validJSONL := `{"type":"message","message":{"id":"msg_01","content":[{"type":"text","text":"Hello, world!"}],"role":"assistant"}}`

	buf := &bytes.Buffer{}
	f := NewStreamFormatter(config.OutputStyleDefault, buf)

	f.ProcessLine(validJSONL)

	// Line number should increment for valid parsed messages
	assert.Equal(t, 1, f.lineNum)
}

func TestStreamFormatter_LineNumberIncrement(t *testing.T) {
	t.Parallel()

	lines := []string{
		`{"type":"message","message":{"id":"1","content":[{"type":"text","text":"Line 1"}],"role":"assistant"}}`,
		`{"type":"message","message":{"id":"2","content":[{"type":"text","text":"Line 2"}],"role":"assistant"}}`,
		`{"type":"message","message":{"id":"3","content":[{"type":"text","text":"Line 3"}],"role":"assistant"}}`,
	}

	buf := &bytes.Buffer{}
	f := NewStreamFormatter(config.OutputStyleDefault, buf)

	for _, line := range lines {
		f.ProcessLine(line)
	}

	assert.Equal(t, 3, f.lineNum)
}

func TestFormatterWriter_Write(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		writes     []string
		wantLines  int
		wantOutput bool
	}{
		"single complete line": {
			writes:     []string{"line one\n"},
			wantLines:  1,
			wantOutput: true,
		},
		"multiple writes for one line": {
			writes:     []string{"line ", "one\n"},
			wantLines:  1,
			wantOutput: true,
		},
		"multiple lines in one write": {
			writes:     []string{"line one\nline two\n"},
			wantLines:  2,
			wantOutput: true,
		},
		"partial line (no newline)": {
			writes:     []string{"partial"},
			wantLines:  0,
			wantOutput: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			w := NewFormatterWriter(config.OutputStyleRaw, buf)

			for _, write := range tt.writes {
				n, err := w.Write([]byte(write))
				require.NoError(t, err)
				assert.Equal(t, len(write), n)
			}

			if tt.wantOutput {
				assert.NotEmpty(t, buf.String())
			}
		})
	}
}

func TestFormatterWriter_Flush(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	w := NewFormatterWriter(config.OutputStyleRaw, buf)

	// Write partial line without newline
	_, err := w.Write([]byte("partial line"))
	require.NoError(t, err)
	assert.Empty(t, buf.String()) // Not flushed yet

	// Flush should process remaining buffer
	w.Flush()
	assert.Contains(t, buf.String(), "partial line")
}

func TestIndexOfNewline(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input []byte
		want  int
	}{
		"newline at start":  {input: []byte("\nrest"), want: 0},
		"newline at end":    {input: []byte("line\n"), want: 4},
		"newline in middle": {input: []byte("ab\ncd"), want: 2},
		"no newline":        {input: []byte("no newline"), want: -1},
		"empty input":       {input: []byte{}, want: -1},
		"multiple newlines": {input: []byte("a\nb\nc"), want: 1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := indexOfNewline(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
