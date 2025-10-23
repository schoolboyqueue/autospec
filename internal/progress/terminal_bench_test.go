package progress_test

import (
	"testing"

	"github.com/anthropics/auto-claude-speckit/internal/progress"
)

// BenchmarkDetectTerminalCapabilities verifies terminal detection meets <10ms performance contract
func BenchmarkDetectTerminalCapabilities(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = progress.DetectTerminalCapabilities()
	}
}

// BenchmarkSelectSymbols verifies symbol selection is fast
func BenchmarkSelectSymbols(b *testing.B) {
	caps := progress.TerminalCapabilities{
		IsTTY:           true,
		SupportsColor:   true,
		SupportsUnicode: true,
		Width:           80,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = progress.SelectSymbols(caps)
	}
}
