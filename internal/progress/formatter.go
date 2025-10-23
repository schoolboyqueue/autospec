package progress

import (
	"fmt"
	"strings"
)

// formatPhaseCounter returns the [N/Total] phase counter string
func formatPhaseCounter(number, total int) string {
	return fmt.Sprintf("[%d/%d]", number, total)
}

// buildPhaseMessage constructs the complete phase message with optional retry info
func buildPhaseMessage(phase PhaseInfo, action string) string {
	counter := formatPhaseCounter(phase.Number, phase.TotalPhases)
	msg := fmt.Sprintf("%s %s %s phase", counter, action, capitalize(phase.Name))

	if phase.RetryCount > 0 {
		msg += fmt.Sprintf(" (retry %d/%d)", phase.RetryCount+1, phase.MaxRetries)
	}

	return msg
}

// capitalize returns the string with the first letter capitalized
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// checkmark returns the appropriate checkmark symbol
func checkmark(symbols ProgressSymbols, supportsColor bool) string {
	mark := symbols.Checkmark
	if supportsColor && symbols.Checkmark == "✓" {
		mark = "\033[32m" + mark + "\033[0m" // Green
	}
	return mark
}

// failureMark returns the appropriate failure symbol
func failureMark(symbols ProgressSymbols, supportsColor bool) string {
	mark := symbols.Failure
	if supportsColor && symbols.Failure == "✗" {
		mark = "\033[31m" + mark + "\033[0m" // Red
	}
	return mark
}
