package config

import (
	"fmt"
	"strings"
)

// OutputStyle represents the output formatting style for stream-json display.
// Maps to cclean display styles, plus "raw" which bypasses formatting entirely.
type OutputStyle string

// Valid output style values
const (
	// OutputStyleDefault uses cclean's default style with box-drawing characters and colors.
	OutputStyleDefault OutputStyle = "default"
	// OutputStyleCompact uses cclean's compact style with single-line summaries.
	OutputStyleCompact OutputStyle = "compact"
	// OutputStyleMinimal uses cclean's minimal style with reduced visual output.
	OutputStyleMinimal OutputStyle = "minimal"
	// OutputStylePlain uses cclean's plain style without colors (suitable for piping/files).
	OutputStylePlain OutputStyle = "plain"
	// OutputStyleRaw bypasses all formatting, outputting raw JSONL as-is.
	OutputStyleRaw OutputStyle = "raw"
)

// validOutputStyles contains all valid output style values for quick lookup.
var validOutputStyles = map[OutputStyle]bool{
	OutputStyleDefault: true,
	OutputStyleCompact: true,
	OutputStyleMinimal: true,
	OutputStylePlain:   true,
	OutputStyleRaw:     true,
}

// ValidOutputStyleNames returns a sorted list of valid style names for display.
func ValidOutputStyleNames() []string {
	return []string{"default", "compact", "minimal", "plain", "raw"}
}

// ValidateOutputStyle validates that the given string is a valid output style.
// Returns nil if valid, or an error with the list of valid options if invalid.
func ValidateOutputStyle(style string) error {
	if style == "" {
		return nil // Empty string defaults to "default" during resolution
	}

	normalized := OutputStyle(strings.ToLower(strings.TrimSpace(style)))
	if !validOutputStyles[normalized] {
		return fmt.Errorf(
			"invalid output_style %q; valid options: %s",
			style,
			strings.Join(ValidOutputStyleNames(), ", "),
		)
	}
	return nil
}

// NormalizeOutputStyle normalizes and validates a style string, returning the
// canonical OutputStyle value. Returns OutputStyleDefault if empty.
func NormalizeOutputStyle(style string) (OutputStyle, error) {
	if style == "" {
		return OutputStyleDefault, nil
	}

	normalized := OutputStyle(strings.ToLower(strings.TrimSpace(style)))
	if !validOutputStyles[normalized] {
		return "", fmt.Errorf(
			"invalid output_style %q; valid options: %s",
			style,
			strings.Join(ValidOutputStyleNames(), ", "),
		)
	}
	return normalized, nil
}

// IsRaw returns true if this style bypasses all formatting.
func (s OutputStyle) IsRaw() bool {
	return s == OutputStyleRaw
}

// String implements fmt.Stringer for OutputStyle.
func (s OutputStyle) String() string {
	return string(s)
}
