# OpenCode Agent Manual Testing Plan

Feature: 077-opencode-agent-preset

## Prerequisites

```bash
# Verify opencode is installed
opencode --version

# Verify autospec is built
make build && ./bin/autospec version
```

## Test Repository Setup

Test repo: `~/repos/opencode-test`

## Test Cases

### 1. Init Command (--ai flag)

```bash
# TC-1a: Init with opencode only
autospec init --ai opencode

# Expected:
# - .opencode/command/autospec.*.md files created
# - opencode.json created with "autospec *": "allow"

# TC-1b: Init with multiple agents
autospec init --ai claude,opencode

# Expected:
# - .claude/commands/autospec.*.md files created
# - .opencode/command/autospec.*.md files created
# - opencode.json created
```

### 2. Doctor Command

```bash
# TC-2: Verify opencode detection
autospec doctor

# Expected:
# - opencode: installed (version shown)
```

### 3. Specify Stage

```bash
# TC-3: Run specify with opencode
AUTOSPEC_AGENT_PRESET=opencode autospec specify "add hello endpoint"

# Expected:
# - OpenCode invoked with: opencode run "<prompt>" --command autospec.specify
# - spec.yaml created in specs/<branch>/
```

### 4. Plan Stage

```bash
# TC-4: Run plan with opencode
AUTOSPEC_AGENT_PRESET=opencode autospec plan

# Expected:
# - OpenCode invoked with: opencode run "<prompt>" --command autospec.plan
# - plan.yaml created
```

### 5. Tasks Stage

```bash
# TC-5: Run tasks with opencode
AUTOSPEC_AGENT_PRESET=opencode autospec tasks

# Expected:
# - OpenCode invoked with: opencode run "<prompt>" --command autospec.tasks
# - tasks.yaml created
```

### 6. Implement Stage

```bash
# TC-6a: Run implement with opencode (full)
AUTOSPEC_AGENT_PRESET=opencode autospec implement

# TC-6b: Run implement with --phases
AUTOSPEC_AGENT_PRESET=opencode autospec implement --phases

# TC-6c: Run implement with --tasks
AUTOSPEC_AGENT_PRESET=opencode autospec implement --tasks
```

### 7. Full Pipeline (run command)

```bash
# TC-7: Run full pipeline
AUTOSPEC_AGENT_PRESET=opencode autospec run -a "add goodbye endpoint"
```

### 8. Prep Command

```bash
# TC-8: Run prep (specify → plan → tasks)
AUTOSPEC_AGENT_PRESET=opencode autospec prep "add status endpoint"
```

### 9. Interactive Commands

```bash
# TC-9a: Clarify (uses --prompt mode)
AUTOSPEC_AGENT_PRESET=opencode autospec clarify

# TC-9b: Analyze
AUTOSPEC_AGENT_PRESET=opencode autospec analyze
```

### 10. Retry System

```bash
# TC-10: Test retry with validation failure
# Create invalid spec.yaml, run plan, verify retry prompt injection
```

### 11. Timeout Configuration

```bash
# TC-11: Test timeout
AUTOSPEC_AGENT_PRESET=opencode AUTOSPEC_TIMEOUT=30s autospec specify "test"

# Expected: Should timeout after 30s if not complete
```

### 12. Config Flag Override

```bash
# TC-12: Use --agent flag instead of env var
autospec specify "test" --agent opencode
```

## Validation Checklist

- [ ] TC-1a: Init with opencode creates correct files
- [ ] TC-1b: Init with multiple agents works
- [ ] TC-2: Doctor detects opencode
- [ ] TC-3: Specify works with opencode
- [ ] TC-4: Plan works with opencode
- [ ] TC-5: Tasks works with opencode
- [ ] TC-6a: Implement (full) works
- [ ] TC-6b: Implement --phases works
- [ ] TC-6c: Implement --tasks works
- [ ] TC-7: Run full pipeline works
- [ ] TC-8: Prep command works
- [ ] TC-9a: Clarify (interactive) works
- [ ] TC-9b: Analyze works
- [ ] TC-10: Retry injects errors correctly
- [ ] TC-11: Timeout respected
- [ ] TC-12: --agent flag works

## Issues Found

| TC | Issue | Status |
|----|-------|--------|
| | | |

