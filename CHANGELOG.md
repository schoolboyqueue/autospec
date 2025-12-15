# Changelog

All notable changes to Auto Claude SpecKit will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Structured error handling with actionable remediation guidance
- Command help output with usage examples for better user guidance
- `--dry-run` flag for commands to preview actions without execution
- Required constitution checks before command execution
- Unit tests for error handling and command help output validation
- `update-task` command for managing task status in tasks.yaml during implementation
- Optional phase commands: `constitution`, `clarify`, `checklist`, `analyze`
- Phase selection flags for the `run` command (`-s`, `-p`, `-t`, `-i`, `-a`)
- Optional phase flags: `-n` (constitution), `-r` (clarify), `-l` (checklist), `-z` (analyze)
- Flexible phase workflow with canonical execution order
- `run` command for flexible execution of workflow phases
- Renamed `full` command to `all` for better clarity
- Artifact dependency checks in preflight validation
- `PhaseConfig` for managing execution phases and dependencies
- Helper scripts for YAML workflow automation
- `skip_confirmations` config option and `AUTOSPEC_YES` environment variable
- Implementation command for executing tasks defined in tasks.yaml
- YAML-structured output format for specifications (spec.yaml, plan.yaml, tasks.yaml)
- Comprehensive documentation (ARCHITECTURE.md, OVERVIEW.md, QUICKSTART.md, REFERENCE.md)
- GitHub issue templates (bug report, feature request)
- Pull request template
- Shell completion support
- Troubleshooting guide

### Changed

- Updated command descriptions and help text throughout CLI
- Checklist structure now uses 'description' and 'status' fields
- Preflight checks now focus on YAML artifacts
- Renamed speckit commands to autospec for consistency

### Fixed

- Constitution requirement checks across all commands
- Task status tracking during implementation
- Artifact dependency validation

## [0.2.0] - 2025-01-15

### Added

- Workflow progress indicators with spinners
- Command execution timeout support
- Timeout configuration via `AUTOSPEC_TIMEOUT` environment variable
- Exit code 5 for timeout errors
- Configurable timeout in config files (0 for infinite, 1-604800 seconds)

### Changed

- Enhanced workflow orchestration with better error handling
- Improved phase execution with clearer status messages

## [0.1.0] - 2025-01-01

### Added

- Initial Go binary implementation
- CLI commands: `workflow`, `specify`, `plan`, `tasks`, `implement`, `status`, `init`, `config`, `doctor`, `version`
- Cobra-based command structure with global flags
- Workflow orchestration (specify -> plan -> tasks -> implement)
- Hierarchical configuration system using koanf
- Configuration sources: environment variables, local config, global config, defaults
- Retry management with persistent state tracking
- Atomic file writes for retry state consistency
- Validation system with <10ms performance contract
- Spec detection from git branch or most recently modified directory
- Git integration helpers
- Pre-flight dependency checks (claude, specify CLIs)
- Claude execution modes: CLI, API, and custom command
- Custom command support with `{{PROMPT}}` placeholder
- Exit code conventions for programmatic use
- Cross-platform builds (Linux, macOS, Windows)

### Changed

- Migrated from bash scripts to Go binary
- Replaced manual validation with automated checks

### Deprecated

- Legacy bash scripts in `scripts/` (scheduled for removal)
- Bats tests in `tests/` (being replaced by Go tests)

[Unreleased]: https://github.com/ariel-frischer/autospec/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/ariel-frischer/autospec/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ariel-frischer/autospec/releases/tag/v0.1.0
