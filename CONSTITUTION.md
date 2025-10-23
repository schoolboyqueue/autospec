# Auto Claude SpecKit - Project Constitution

**Version**: 1.0.0
**Last Updated**: 2025-10-22
**Status**: Active

---

## Preamble

This constitution establishes the foundational principles, values, and guidelines for the Auto Claude SpecKit project. It serves as the authoritative reference for decision-making, development practices, and project evolution.

---

## I. Mission & Vision

### Mission Statement
To provide a reliable, performant, and user-friendly tool that automates SpecKit workflow validation for Claude Code, enabling developers to build features with confidence and consistency.

### Vision
A cross-platform, zero-dependency CLI tool that becomes the standard for SpecKit workflow automation, trusted by developers worldwide for its reliability, simplicity, and extensibility.

### Core Purpose
Eliminate manual validation overhead in SpecKit workflows by providing intelligent automation that ensures completeness, enforces standards, and accelerates feature development.

---

## II. Core Values

### 1. Reliability Above All
- **Principle**: Users must trust that validation is accurate and consistent
- **Implications**:
  - Comprehensive testing (>80% coverage minimum)
  - Parity validation with legacy bash implementation
  - Defensive error handling with clear error messages
  - No silent failures

### 2. User Experience First
- **Principle**: The tool should be intuitive and reduce friction
- **Implications**:
  - Single binary distribution, no complex installation
  - Clear, actionable error messages
  - Fast execution (<5s for full workflows)
  - Progressive disclosure: simple for basic use, powerful for advanced use
  - Documentation that prioritizes clarity over completeness

### 3. Zero Dependencies Promise
- **Principle**: Users should not need to install additional tools
- **Implications**:
  - Pure Go implementation (no bash, jq, grep, sed dependencies)
  - Embedded git operations (no git binary required)
  - Self-contained binary with no external runtime requirements
  - Only exception: Claude CLI itself

### 4. Platform Agnosticism
- **Principle**: Works identically on all major platforms
- **Implications**:
  - Cross-platform file path handling
  - No platform-specific shell assumptions
  - Tested on Linux, macOS, Windows
  - Consistent behavior across environments

### 5. Performance Consciousness
- **Principle**: Validation should be imperceptible to users
- **Implications**:
  - <50ms startup time
  - Sub-second validation for typical workflows
  - Efficient file parsing and processing
  - No unnecessary I/O or computation

### 6. Maintainability & Clarity
- **Principle**: Code should be easy to understand and modify
- **Implications**:
  - Clear, self-documenting code
  - Minimal dependencies (only well-maintained, necessary libraries)
  - Comprehensive inline documentation
  - Architecture that supports extension without modification

---

## III. Design Principles

### 1. Progressive Complexity
- **Simple**: Basic usage requires minimal configuration
- **Configurable**: Advanced users can customize behavior
- **Extensible**: Power users can integrate with custom workflows

**Example**:
```bash
# Simple: Works out of the box
autospec workflow "my feature"

# Configurable: Custom settings
autospec workflow "my feature" --max-retries 5

# Extensible: Custom Claude command template
# Via config.json: custom_claude_cmd with {{PROMPT}} placeholder
```

### 2. Explicit Over Implicit
- Configuration should be explicit and discoverable
- Validation failures provide clear reasons
- No "magic" behavior without user understanding
- Exit codes follow documented conventions

### 3. Fail Fast, Fail Clear
- Detect errors as early as possible
- Provide actionable error messages with solutions
- Include context in error reporting
- Never leave users guessing about what went wrong

### 4. Convention Over Configuration
- Sensible defaults for 90% use cases
- Standard patterns (e.g., `specs/` directory)
- Override capability for special needs
- Auto-detect when possible (e.g., current spec from branch)

### 5. Composability
- Individual commands work standalone
- Commands can be chained programmatically
- JSON output mode for tool integration
- Exit codes enable scripting

---

## IV. Technical Standards

### Code Quality

#### Mandatory
- All code must pass `go vet`
- All code must be formatted with `gofmt`
- All exported functions must have godoc comments
- All errors must be wrapped with context
- No panics in library code (only in main)

#### Strongly Encouraged
- Use static analysis tools (golangci-lint)
- Follow Effective Go guidelines
- Prefer explicit error handling over silent failures
- Use interfaces for testability

### Testing Requirements

#### Coverage Standards
- **Minimum**: 80% code coverage
- **Critical paths**: 100% coverage (validation logic, retry mechanisms)
- **Integration tests**: All CLI commands
- **Platform tests**: CI runs on Linux, macOS, Windows

#### Test Categories
1. **Unit tests**: Individual functions, pure logic
2. **Integration tests**: CLI command execution
3. **Parity tests**: Behavior matches bash version
4. **Performance tests**: Validation speed benchmarks
5. **Platform tests**: Cross-platform compatibility

#### Test Principles
- Tests should be fast (<1s for unit tests)
- Tests should be deterministic (no flaky tests)
- Use table-driven tests for multiple scenarios
- Mock external dependencies (filesystem, git, Claude CLI)

### Documentation Standards

#### Required Documentation
1. **Code comments**: Exported functions and complex logic
2. **README**: Installation, quick start, basic usage
3. **Architecture docs**: High-level design and data flow
4. **Migration guide**: Transitioning from bash version
5. **Contribution guide**: How to contribute effectively

#### Documentation Principles
- Start with examples, then explain
- Keep examples up-to-date with code
- Write for beginners, provide depth for experts
- Include troubleshooting for common issues

---

## V. Development Process

### Version Control

#### Branching Strategy
- **main**: Stable, production-ready code
- **develop**: Integration branch for features
- **feature/\***: Individual feature branches
- **hotfix/\***: Critical bug fixes for production

#### Commit Standards
- Follow [Conventional Commits](https://www.conventionalcommits.org/)
- Format: `type(scope): description`
- Types: feat, fix, docs, test, refactor, perf, chore
- Include issue references when applicable

**Examples**:
```
feat(validator): add support for nested task phases
fix(cli): handle spaces in file paths correctly
docs(readme): update installation instructions
test(implement): add parity tests for task counting
```

### Pull Request Requirements

#### Before Opening PR
- [ ] All tests pass locally
- [ ] Code is formatted (`make fmt`)
- [ ] No linting errors (`make lint`)
- [ ] Documentation updated if needed
- [ ] Changelog entry added (if user-facing)

#### PR Description Must Include
- What: What does this change do?
- Why: Why is this change necessary?
- How: How does it work? (for non-trivial changes)
- Testing: How was it tested?

#### Review Criteria
- **Functionality**: Does it work as intended?
- **Tests**: Are tests comprehensive?
- **Performance**: Does it maintain performance standards?
- **Documentation**: Is it documented appropriately?
- **Code quality**: Is it maintainable and clear?

### Release Process

#### Version Numbering
- Semantic Versioning (SemVer 2.0.0)
- Format: MAJOR.MINOR.PATCH
- Pre-release: MAJOR.MINOR.PATCH-beta.N

#### Release Criteria
- All tests passing on all platforms
- Documentation reviewed and updated
- Changelog complete and accurate
- Performance benchmarks acceptable
- Breaking changes clearly documented

#### Release Steps
1. Update version in code and docs
2. Update CHANGELOG.md
3. Create and push version tag
4. GitHub Actions builds and releases
5. Verify release artifacts
6. Announce release (if significant)

---

## VI. Governance

### Decision-Making

#### Levels of Decision
1. **Trivial**: Any contributor (e.g., typo fixes, doc clarifications)
2. **Minor**: Maintainer approval (e.g., bug fixes, small features)
3. **Major**: Consensus discussion (e.g., architecture changes, breaking changes)
4. **Constitutional**: Community RFC (e.g., changes to core values, mission)

#### RFC Process (for Major/Constitutional decisions)
1. Open issue with "RFC:" prefix
2. Describe problem, proposed solution, alternatives
3. Minimum 7-day discussion period
4. Address feedback, update proposal
5. Final decision by maintainer(s)
6. Document decision in ADR (Architecture Decision Record)

### Maintainer Responsibilities

#### Primary Maintainer Duties
- Review and merge pull requests
- Ensure CI/CD pipeline health
- Triage and respond to issues
- Maintain roadmap and milestone planning
- Enforce code quality and testing standards
- Release management

#### Maintainer Principles
- **Responsive**: Acknowledge issues/PRs within 3 days
- **Transparent**: Explain decisions clearly
- **Welcoming**: Foster inclusive community
- **Pragmatic**: Balance idealism with practical constraints

### Community Guidelines

#### Expected Behavior
- Respectful and professional communication
- Constructive feedback and criticism
- Collaborative problem-solving
- Recognition of contributions

#### Unacceptable Behavior
- Harassment or discriminatory language
- Unconstructive criticism or hostility
- Spam or off-topic discussions
- Violations of privacy or security

---

## VII. Stability Guarantees

### Semantic Versioning Commitments

#### MAJOR version (X.0.0)
Breaking changes requiring user action:
- CLI command structure changes
- Configuration format changes
- Exit code convention changes
- Removal of features

#### MINOR version (0.X.0)
Backward-compatible additions:
- New commands or features
- New configuration options
- Performance improvements
- Enhanced error messages

#### PATCH version (0.0.X)
Backward-compatible fixes:
- Bug fixes
- Documentation corrections
- Security patches
- Minor performance tweaks

### Deprecation Policy

#### Process
1. Mark feature as deprecated in documentation
2. Add deprecation warning (if possible)
3. Announce in release notes and changelog
4. Maintain for at least 2 MINOR versions
5. Remove in next MAJOR version

#### Example Timeline
- v1.2.0: Deprecate feature X, add warning
- v1.3.0: Feature X still works with warning
- v1.4.0: Feature X still works with warning
- v2.0.0: Feature X removed

### Configuration Compatibility

#### Guarantees
- Configuration files remain compatible within MAJOR version
- Auto-migration provided for breaking config changes
- Clear migration documentation for major versions
- Validation of config files with helpful error messages

---

## VIII. Security

### Security Policy

#### Vulnerability Reporting
- Report security issues privately to maintainer
- Do not open public issues for security vulnerabilities
- Allow 90 days for fix before public disclosure
- Credit reporters in security advisories

#### Security Standards
- No secrets in code or logs
- Validate all user input
- Use secure defaults
- Follow principle of least privilege
- Regular dependency updates for security patches

### Supply Chain Security

#### Dependencies
- Minimize third-party dependencies
- Only use well-maintained, trusted libraries
- Regular security audits of dependencies
- Automated dependency update monitoring
- Verify checksums for releases

---

## IX. Roadmap Principles

### Feature Prioritization

#### Priority Framework
1. **Critical**: Bugs blocking core functionality
2. **High**: Features needed for v1.0 parity
3. **Medium**: Quality-of-life improvements
4. **Low**: Nice-to-have enhancements

#### Decision Criteria
- User impact (how many users benefit?)
- Alignment with core values
- Implementation complexity
- Maintenance burden
- Community demand

### Scope Management

#### In Scope
- SpecKit workflow automation and validation
- Cross-platform CLI tool
- Hook integration with Claude Code
- Configuration management
- State tracking (retry counts, etc.)

#### Out of Scope
- IDE plugins (may be separate project)
- Cloud sync services (may be separate project)
- General-purpose task management
- Features unrelated to SpecKit workflows

### Innovation vs. Stability

#### Stability Phase (v1.x)
- Focus on reliability and performance
- Conservative about new features
- Prioritize bug fixes and polish
- Maintain backward compatibility

#### Innovation Phase (v2.x planning)
- Explore new capabilities
- Consider architectural improvements
- Gather community feedback
- Prototype experimental features

---

## X. Migration from Bash Version

### Parity Requirements

#### Must-Have Parity
- All validation logic produces identical results
- All CLI commands have equivalent functionality
- All configuration options are supported
- Exit codes match bash version conventions

#### Nice-to-Have Improvements
- Better error messages
- Faster performance
- Additional features (that don't break parity)

### Transition Support

#### Commitment
- Maintain bash version documentation in archive
- Provide migration guide
- Support both versions during transition (v0.x cycle)
- Auto-detect and migrate legacy configurations

#### Migration Timeline
- v0.1.0-beta: Basic Go CLI available
- v0.5.0-beta: Feature parity achieved
- v1.0.0: Stable Go release, bash version archived
- v1.x: Go version only, bash version unsupported

---

## XI. Amendments

### Amendment Process

This constitution can be amended through the RFC process:

1. Open RFC issue proposing amendment
2. Discuss for minimum 14 days (extended period for constitutional changes)
3. Incorporate community feedback
4. Final decision by primary maintainer(s)
5. Update constitution with version bump
6. Document rationale in amendment history

### Amendment History

- **v1.0.0** (2025-10-22): Initial constitution established

---

## XII. Acknowledgments

### Inspiration
This project builds upon:
- Claude Code's SpecKit feature development methodology
- Best practices from Go CLI tools (cobra, viper)
- Lessons learned from bash script automation

### Contributors
All contributors are recognized in CONTRIBUTORS.md and release notes.

---

## Conclusion

This constitution establishes the foundation for Auto Claude SpecKit's development. It balances ambition with pragmatism, innovation with stability, and simplicity with power.

**Core Tenets** (in priority order):
1. Reliability above all
2. User experience first
3. Zero dependencies promise
4. Platform agnosticism
5. Performance consciousness
6. Maintainability and clarity

When in doubt, refer to these tenets for guidance.

---

**Signed**: Auto Claude SpecKit Maintainers
**Date**: 2025-10-22
**Version**: 1.0.0
