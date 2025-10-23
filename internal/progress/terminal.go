package progress

import (
	"os"

	"golang.org/x/term"
)

// DetectTerminalCapabilities detects terminal features and returns capabilities
func DetectTerminalCapabilities() TerminalCapabilities {
	// Check if stdout is a terminal
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	// Check environment variables
	noColor := os.Getenv("NO_COLOR") != ""
	forceASCII := os.Getenv("AUTOSPEC_ASCII") == "1"

	// Get terminal width
	width := 0
	if isTTY {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			width = w
		}
	}

	return TerminalCapabilities{
		IsTTY:           isTTY,
		SupportsColor:   isTTY && !noColor,
		SupportsUnicode: isTTY && !forceASCII,
		Width:           width,
	}
}

// SelectSymbols returns the appropriate symbol set based on terminal capabilities
func SelectSymbols(caps TerminalCapabilities) ProgressSymbols {
	if caps.SupportsUnicode {
		return ProgressSymbols{
			Checkmark:  "✓",
			Failure:    "✗",
			SpinnerSet: 14, // Unicode dots: ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
		}
	}

	return ProgressSymbols{
		Checkmark:  "[OK]",
		Failure:    "[FAIL]",
		SpinnerSet: 9, // ASCII: | / - \
	}
}
