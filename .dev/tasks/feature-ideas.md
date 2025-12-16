# Feature Ideas & Improvements

## Implementation Phase Improvements

### Smart Retry Logic for Incomplete Tasks

**Problem**: When Claude stops mid-implementation (context limits, attention issues), the current retry mechanism doesn't provide targeted guidance.

**Proposed Solution**:
```go
// After Claude "finishes" but validation fails:
if !stats.IsComplete() {
    remainingIDs := getRemainingTaskIDs(stats)
    retryPrompt := fmt.Sprintf("Continue implementation. Remaining tasks: %s",
        strings.Join(remainingIDs, ", "))
    // Trigger retry with this specific prompt
}
```

**Benefits**:
- Claude gets explicit list of remaining tasks
- More focused continuation rather than starting from scratch
- Could include task titles for better context

---

### Blocked Task Detection

**Problem**: Retrying blocked tasks is wasteful - they need human intervention.

**Proposed Solution**:
```go
if stats.BlockedTasks > 0 {
    blockedIDs := getBlockedTaskIDs(stats)
    return &RetryDecision{
        ShouldRetry: false,
        Reason:      "blocked tasks require human intervention",
        BlockedTasks: blockedIDs,
        Suggestion:  "Review blocked tasks and update their status before retrying",
    }
}
```

**Decision Matrix**:

| Scenario | Action |
|----------|--------|
| Tasks "Pending" or "InProgress" | Auto-retry with continuation prompt |
| Tasks "Blocked" | Stop and ask human for input |
| Claude errors/timeout | Retry up to max_retries |
| All tasks "Completed" | Success |

---

### Task Progress Persistence

**Problem**: If autospec crashes or user kills the process, there's no way to resume exactly where Claude left off.

**Proposed Solution**:
- Persist task progress more frequently during implementation
- Add `--resume` flag that reads task status and generates a "continue from here" prompt
- Store last-executed task ID in retry state

---

### Continuation Prompt Generation

**Problem**: When retrying, Claude may not understand what was already done.

**Proposed Solution**:
```go
func GenerateContinuationPrompt(stats *TaskStats) string {
    var prompt strings.Builder
    prompt.WriteString("Resume implementation from where you left off.\n\n")

    prompt.WriteString("Completed tasks (DO NOT redo these):\n")
    for _, id := range getCompletedTaskIDs(stats) {
        prompt.WriteString(fmt.Sprintf("- %s\n", id))
    }

    prompt.WriteString("\nRemaining tasks:\n")
    for _, id := range getRemainingTaskIDs(stats) {
        prompt.WriteString(fmt.Sprintf("- %s\n", id))
    }

    if stats.InProgressTasks > 0 {
        prompt.WriteString("\nNote: Task currently in-progress - verify completion before moving on.\n")
    }

    return prompt.String()
}
```

---

## Other Ideas

### Phase-Level Checkpoints
- Save checkpoint after each completed phase
- Allow `autospec implement --from-phase=5` to resume from specific phase

### Max Context Warning
- Track approximate token usage during implementation
- Warn user when approaching context limits
- Suggest breaking into smaller batches

### Parallel Task Execution
- For tasks marked `parallel: true`, consider spawning multiple Claude sessions
- Would require careful coordination and merging of changes

---

## Priority

1. **High**: Smart retry with remaining task list (immediate value)
2. **Medium**: Blocked task detection (prevents wasted retries)
3. **Low**: Progress persistence / checkpoints (nice to have)
