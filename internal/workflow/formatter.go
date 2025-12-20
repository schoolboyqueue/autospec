package workflow

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/ariel-frischer/autospec/internal/config"
	"github.com/ariel-frischer/claude-clean/display"
	"github.com/ariel-frischer/claude-clean/parser"
)

// StreamFormatter handles line-by-line formatting of Claude's stream-json output
// using the cclean library for parsing and display.
type StreamFormatter struct {
	style   config.OutputStyle
	writer  io.Writer
	lineNum int
	config  *display.Config
}

// NewStreamFormatter creates a new StreamFormatter for the given style and output writer.
// The style is mapped to the corresponding cclean display configuration.
func NewStreamFormatter(style config.OutputStyle, writer io.Writer) *StreamFormatter {
	return &StreamFormatter{
		style:   style,
		writer:  writer,
		lineNum: 0,
		config:  mapStyleToConfig(style),
	}
}

// mapStyleToConfig converts an OutputStyle to a cclean display.Config.
// Raw style is handled separately (bypasses formatting entirely).
func mapStyleToConfig(style config.OutputStyle) *display.Config {
	cfg := &display.Config{
		Verbose:     false,
		ShowLineNum: false,
	}

	switch style {
	case config.OutputStyleCompact:
		cfg.Style = display.StyleCompact
	case config.OutputStyleMinimal:
		cfg.Style = display.StyleMinimal
	case config.OutputStylePlain:
		cfg.Style = display.StylePlain
	case config.OutputStyleRaw:
		// Raw doesn't use display at all, but provide a default for safety
		cfg.Style = display.StyleDefault
	default:
		cfg.Style = display.StyleDefault
	}

	return cfg
}

// ProcessLine parses and formats a single JSONL line from Claude's stream-json output.
// If parsing fails, the raw line is returned unchanged.
// Empty lines are silently skipped.
func (f *StreamFormatter) ProcessLine(line string) {
	// Skip empty lines
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return
	}

	// Raw mode: pass through without any processing
	if f.style.IsRaw() {
		f.writeRaw(line)
		return
	}

	// Try to parse as JSON
	var msg parser.StreamMessage
	if err := json.Unmarshal([]byte(trimmed), &msg); err != nil {
		// Parse error: fall back to raw output
		f.writeRaw(line)
		return
	}

	f.lineNum++
	f.formatAndWrite(&msg)
}

// formatAndWrite displays a parsed message using cclean's display package.
// System reminders are stripped for non-raw styles.
func (f *StreamFormatter) formatAndWrite(msg *parser.StreamMessage) {
	// Strip system reminders from message content
	stripSystemReminders(msg)

	// Use cclean's display to format and write the message
	display.DisplayMessage(msg, f.lineNum, f.config)
}

// stripSystemReminders removes system reminder tags from all text content blocks.
func stripSystemReminders(msg *parser.StreamMessage) {
	if msg == nil || msg.Message == nil {
		return
	}

	for i := range msg.Message.Content {
		if msg.Message.Content[i].Type == "text" {
			msg.Message.Content[i].Text = parser.StripSystemReminders(msg.Message.Content[i].Text)
		}
	}
}

// writeRaw writes a line directly to the output writer without formatting.
func (f *StreamFormatter) writeRaw(line string) {
	f.writer.Write([]byte(line))
	if !strings.HasSuffix(line, "\n") {
		f.writer.Write([]byte("\n"))
	}
}

// FormatReader reads lines from an io.Reader and formats each line.
// This is a convenience method for processing an entire stream.
func (f *StreamFormatter) FormatReader(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	// Increase buffer size to handle long JSON lines (64KB default may not be enough)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // Max 1MB per line

	for scanner.Scan() {
		f.ProcessLine(scanner.Text())
	}

	return scanner.Err()
}

// FormatterWriter wraps a StreamFormatter to implement io.Writer.
// Each Write call is processed as a potential multi-line input.
type FormatterWriter struct {
	formatter *StreamFormatter
	buffer    []byte
}

// NewFormatterWriter creates an io.Writer that formats stream-json output.
func NewFormatterWriter(style config.OutputStyle, output io.Writer) *FormatterWriter {
	return &FormatterWriter{
		formatter: NewStreamFormatter(style, output),
		buffer:    make([]byte, 0, 4096),
	}
}

// Write implements io.Writer. It buffers partial lines and processes complete lines.
func (w *FormatterWriter) Write(p []byte) (n int, err error) {
	w.buffer = append(w.buffer, p...)

	// Process complete lines
	for {
		idx := indexOfNewline(w.buffer)
		if idx < 0 {
			break
		}

		line := string(w.buffer[:idx])
		w.buffer = w.buffer[idx+1:]
		w.formatter.ProcessLine(line)
	}

	return len(p), nil
}

// Flush processes any remaining data in the buffer.
func (w *FormatterWriter) Flush() {
	if len(w.buffer) > 0 {
		w.formatter.ProcessLine(string(w.buffer))
		w.buffer = w.buffer[:0]
	}
}

// indexOfNewline finds the index of the first newline character in b.
// Returns -1 if no newline is found.
func indexOfNewline(b []byte) int {
	for i, c := range b {
		if c == '\n' {
			return i
		}
	}
	return -1
}
