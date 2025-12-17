// Package lifecycle provides wrapper functions for CLI command and workflow
// stage execution. It handles timing and notification dispatch, eliminating
// boilerplate code across CLI commands.
//
// The lifecycle package is intentionally minimal: no event bus, no goroutines,
// no external dependencies. Each wrapper function captures start time, executes
// the provided function, calculates duration, and calls the appropriate
// notification method.
package lifecycle

import (
	"context"
	"time"
)

// Run wraps command execution with timing and notification dispatch.
// It captures the start time, executes fn, calculates duration, and calls
// handler.OnCommandComplete with the results.
//
// If handler is nil, fn is still executed but no notification is sent.
// Handler panics are recovered to ensure command completion is not affected.
// The original error from fn is always returned unchanged.
func Run(handler NotificationHandler, name string, fn func() error) error {
	start := time.Now()
	fnErr := fn()
	duration := time.Since(start)

	notifyCommandComplete(handler, name, fnErr == nil, duration)

	return fnErr
}

// RunWithContext wraps context-aware command execution.
// If the context is already cancelled, returns context.Canceled immediately
// without executing fn. Otherwise behaves like Run.
//
// The notification is sent regardless of whether fn was executed or cancelled.
func RunWithContext(ctx context.Context, handler NotificationHandler, name string, fn func(context.Context) error) error {
	start := time.Now()

	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		duration := time.Since(start)
		notifyCommandComplete(handler, name, false, duration)
		return err
	}

	fnErr := fn(ctx)
	duration := time.Since(start)

	notifyCommandComplete(handler, name, fnErr == nil, duration)

	return fnErr
}

// RunStage wraps workflow stage execution with notification dispatch.
// It executes fn and calls handler.OnStageComplete with the results.
//
// If handler is nil, fn is still executed but no notification is sent.
// Handler panics are recovered to ensure stage completion is not affected.
func RunStage(handler NotificationHandler, name string, fn func() error) error {
	fnErr := fn()
	notifyStageComplete(handler, name, fnErr == nil)
	return fnErr
}

// notifyCommandComplete safely calls OnCommandComplete with panic recovery.
func notifyCommandComplete(handler NotificationHandler, name string, success bool, duration time.Duration) {
	if handler == nil {
		return
	}
	defer func() { _ = recover() }()
	handler.OnCommandComplete(name, success, duration)
}

// notifyStageComplete safely calls OnStageComplete with panic recovery.
func notifyStageComplete(handler NotificationHandler, name string, success bool) {
	if handler == nil {
		return
	}
	defer func() { _ = recover() }()
	handler.OnStageComplete(name, success)
}
