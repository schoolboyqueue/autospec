package progress

import (
	"fmt"
	"strings"
)

// formatStageCounter returns the [N/Total] stage counter string
func formatStageCounter(number, total int) string {
	return fmt.Sprintf("[%d/%d]", number, total)
}

// buildStageMessage constructs the complete stage message with optional retry info
func buildStageMessage(stage StageInfo, action string) string {
	counter := formatStageCounter(stage.Number, stage.TotalStages)
	msg := fmt.Sprintf("%s %s %s stage", counter, action, capitalize(stage.Name))

	if stage.RetryCount > 0 {
		msg += fmt.Sprintf(" (retry %d/%d)", stage.RetryCount+1, stage.MaxRetries)
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
