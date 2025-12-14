# Feature Specification: Flexible Phase Workflow

**Feature Branch**: `008-flexible-phase-workflow`
**Created**: 2025-12-13
**Status**: Draft
**Input**: User description: "Allow users to run custom combinations of SpecKit phases in any order with safety warnings"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Custom Phase Combination (Priority: P1)

As a developer working on a feature with an existing spec, I want to run only the plan and implement phases without regenerating the spec, so that I can resume work efficiently without redundant steps.

**Why this priority**: This is the core value proposition - enabling flexible phase combinations saves time for experienced users who know exactly which phases they need.

**Independent Test**: Can be fully tested by running the tool with `-pi` flags on an existing spec branch and verifying both plan and implement phases execute in order.

**Acceptance Scenarios**:

1. **Given** I am on a branch with an existing spec.md file, **When** I run the command with plan and implement flags only, **Then** the system executes plan followed by implement phases, skipping the specify phase.

2. **Given** I specify multiple phases via flags, **When** I run the command, **Then** the phases always execute in the canonical order (specify → plan → tasks → implement) regardless of the order I specified the flags.

3. **Given** I am on a spec branch with all artifacts present, **When** I run a subset of phases, **Then** the system proceeds immediately without any confirmation prompts.

---

### User Story 2 - Run All Phases with Single Flag (Priority: P1)

As a developer starting a new feature, I want a quick way to run all four phases with a single command or flag, so that I can set up a complete workflow with minimal typing.

**Why this priority**: This is the most common workflow use case and must be easy to invoke.

**Independent Test**: Can be fully tested by running the tool with `-a` flag and a feature description, verifying all four phases execute sequentially.

**Acceptance Scenarios**:

1. **Given** I have a feature description, **When** I run the command with the "all phases" flag, **Then** the system executes specify, plan, tasks, and implement phases in sequence.

2. **Given** I run the "all" subcommand with a feature description, **When** the command completes, **Then** the result is identical to running with the "all phases" flag.

---

### User Story 3 - Safety Warnings for Missing Prerequisites (Priority: P2)

As a developer who may accidentally skip required phases, I want the system to warn me when prerequisite artifacts are missing and ask for confirmation, so that I don't waste time running phases that will fail.

**Why this priority**: Prevents user frustration from failed executions due to missing dependencies, providing a safety net without blocking advanced users.

**Independent Test**: Can be fully tested by attempting to run tasks phase without a plan.md file present and verifying a warning appears with confirmation prompt.

**Acceptance Scenarios**:

1. **Given** I request the tasks phase but plan.md does not exist, **When** I run the command, **Then** the system displays a warning about the missing prerequisite and asks for confirmation before proceeding.

2. **Given** I am shown a prerequisite warning, **When** I confirm with "y", **Then** the system proceeds with execution.

3. **Given** I am shown a prerequisite warning, **When** I decline with "n" or press Enter, **Then** the system aborts without executing any phases.

---

### User Story 4 - Skip Confirmation Prompts (Priority: P2)

As an automation script or experienced user, I want to skip all confirmation prompts, so that I can run commands non-interactively or without interruption.

**Why this priority**: Essential for CI/CD integration and power users who want streamlined execution.

**Independent Test**: Can be fully tested by running with the "yes" flag when prerequisites are missing and verifying no prompt appears.

**Acceptance Scenarios**:

1. **Given** I pass the "skip confirmation" flag, **When** prerequisites are missing, **Then** the system proceeds without showing a confirmation prompt.

2. **Given** I set the "skip confirmations" environment variable, **When** I run a command with missing prerequisites, **Then** the system proceeds without prompts.

3. **Given** I configure "skip confirmations" in my config file, **When** I run commands, **Then** all confirmation prompts are skipped by default.

---

### User Story 5 - Explicit Spec Selection (Priority: P3)

As a developer working across multiple features, I want to explicitly specify which spec to work with, so that I can run phases on a different spec than my current branch suggests.

**Why this priority**: Provides flexibility for advanced workflows where branch naming conventions don't apply.

**Independent Test**: Can be fully tested by passing an explicit spec name flag while on a different branch and verifying the specified spec is used.

**Acceptance Scenarios**:

1. **Given** I am on branch "main", **When** I run a command with an explicit spec name flag, **Then** the system uses the specified spec instead of trying to detect from branch.

2. **Given** I provide a non-existent spec name, **When** I run the command without specify phase, **Then** the system displays an error indicating the spec was not found.

---

### User Story 6 - Branch-Aware Spec Detection (Priority: P3)

As a developer following git workflow conventions, I want the system to automatically detect which spec I'm working on based on my current branch name, so that I don't have to specify it manually.

**Why this priority**: Reduces friction for users following standard naming conventions, making the tool feel intelligent.

**Independent Test**: Can be fully tested by checking out a spec branch and running a phase command without specifying the spec name.

**Acceptance Scenarios**:

1. **Given** I am on branch "007-yaml-output", **When** I run a phase command without specifying a spec, **Then** the system detects and uses spec "007-yaml-output".

2. **Given** I am on a non-spec branch like "main", **When** I run a phase command without the specify phase and without an explicit spec, **Then** the system displays an error with suggestions on how to proceed.

3. **Given** I am in detached HEAD state or outside a git repository, **When** I run a phase command, **Then** the system falls back to using the most recently modified spec directory.

---

### Edge Cases

- What happens when the user provides conflicting flags (e.g., both `-s` requiring a description and no description provided)?
  - System displays an error: "Feature description required when using specify phase"

- What happens when the user specifies no phase flags at all?
  - System shows help/usage information

- What happens when the git repository is in an unusual state (detached HEAD, no git, corrupted)?
  - System falls back to most recent spec directory detection

- What happens when prerequisites are missing but the user confirms to proceed?
  - System attempts execution; individual phase may fail with appropriate error messages

- What happens when all phase flags and the "all" flag are provided together?
  - System treats this as equivalent to "all" - no conflict, phases execute once

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support individual phase flags (`-s`/`--specify`, `-p`/`--plan`, `-t`/`--tasks`, `-i`/`--implement`) that can be combined in any order.

- **FR-002**: System MUST support an "all phases" shortcut (`-a`/`--all`) that enables all four phases.

- **FR-003**: System MUST execute selected phases in canonical order (specify → plan → tasks → implement) regardless of the order flags are specified.

- **FR-004**: System MUST detect the current spec from the git branch name when the branch follows the pattern `NNN-feature-name`.

- **FR-005**: System MUST display a warning and request confirmation when attempting to run a phase without its prerequisite artifacts present.

- **FR-006**: System MUST support skipping confirmation prompts via flag (`-y`/`--yes`), environment variable (`AUTOSPEC_YES`), or configuration setting (`skip_confirmations`).

- **FR-007**: System MUST support explicit spec selection via the `--spec` flag, overriding branch-based detection.

- **FR-008**: System MUST require a feature description argument when the specify phase is included in the execution.

- **FR-009**: System MUST display an error with actionable suggestions when on a non-spec branch without the specify phase and without explicit spec selection.

- **FR-010**: System MUST fall back to the most recently modified spec directory when git branch detection fails (detached HEAD, no git repository).

- **FR-011**: System MUST rename the existing "full" subcommand to "all" for executing all four phases in sequence.

- **FR-012**: System MUST show help/usage when invoked without any phase flags or subcommands.

### Key Entities

- **PhaseConfig**: Represents the user's selected phases (specify, plan, tasks, implement, all) and determines execution order.

- **PreflightResult**: Contains the results of prerequisite checking including detected spec, existing artifacts, warnings, and whether confirmation is needed.

- **Artifact Dependency Map**: Defines the relationship between phases and their required/produced artifacts:
  - specify → creates spec.md
  - plan → requires spec.md, creates plan.md
  - tasks → requires plan.md, creates tasks.md
  - implement → requires tasks.md

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can specify any combination of phases and have them execute in the correct order within the same command invocation.

- **SC-002**: Users can run the complete four-phase workflow with 3 or fewer characters of flags (`-a`).

- **SC-003**: Warning messages for missing prerequisites are displayed in under 1 second after command invocation.

- **SC-004**: Users on spec branches with all artifacts present experience zero confirmation prompts for any phase combination.

- **SC-005**: Users receive actionable error messages with specific suggestions when spec detection fails.

## Assumptions

- Users are familiar with Unix-style short flags that can be combined (e.g., `-spi` similar to `tar -xvf`).
- The existing git integration and spec detection logic can be extended without breaking changes.
- The Cobra CLI framework supports adding persistent flags to the root command that work alongside subcommands.
- Confirmation prompts use simple y/N input on stdin; no complex terminal UI is required.
- The canonical phase order (specify → plan → tasks → implement) is fixed and will not change.
