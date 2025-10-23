# Auto Claude SpecKit - Go Binary Transition Plan

## Overview

Transform the current bash-based validation tool into a single, cross-platform Go binary that provides a professional CLI experience while maintaining all existing functionality.

## Goals

- **Single Binary**: One executable that works across Linux, macOS, Windows
- **Easy Installation**: `go install` or download from releases
- **Simple UX**: `autospec workflow "feature"` - works anywhere
- **No Dependencies**: Embed all scripts, no bash/jq/git dependencies required
- **Maintainable**: Clean Go codebase, easy for contributors
- **Backward Compatible**: Existing validation logic preserved

## Architecture Decision: Hybrid Approach

### Option A: Pure Go Rewrite ❌
Rewrite all bash validation logic in Go.

**Pros**: Native performance, no shell dependencies
**Cons**:
- High effort to port all logic
- Need to rewrite 60+ tests
- Risk of introducing bugs during translation
- Lose bash expertise already invested

### Option B: Go Wrapper + Embedded Bash ❌
Go binary embeds and executes bash scripts.

**Pros**: Reuse existing tested logic
**Cons**:
- Still requires bash/jq/git on user systems
- Platform compatibility issues (Windows)
- Defeats purpose of "single binary"
- Not acceptable for this project

### Option C: Pure Go with Complete Logic Rewrite ✅ **SELECTED**
**100% Go implementation** - rewrite ALL validation logic in Go, no bash execution.

**Pros**:
- True cross-platform support (Linux/macOS/Windows)
- Zero runtime dependencies
- Better performance
- Native Go testing
- Professional, maintainable codebase
- Single binary distribution

**Cons**:
- Initial development effort (2-3 weeks)
- Need to validate parity with bash version

**Decision**: Pure Go implementation. Bash scripts are kept ONLY as reference during porting, then moved to `legacy/` branch and deleted from main.

---

## Project Structure

```
auto-claude-speckit/
├── cmd/
│   └── autospec/
│       └── main.go                    # CLI entry point
├── internal/
│   ├── cli/
│   │   ├── root.go                   # Root command
│   │   ├── init.go                   # autospec init
│   │   ├── workflow.go               # autospec workflow
│   │   ├── specify.go                # autospec specify
│   │   ├── plan.go                   # autospec plan
│   │   ├── tasks.go                  # autospec tasks
│   │   ├── implement.go              # autospec implement
│   │   ├── status.go                 # autospec status
│   │   └── config.go                 # autospec config
│   ├── validator/
│   │   ├── spec.go                   # Spec validation logic
│   │   ├── plan.go                   # Plan validation logic
│   │   ├── tasks.go                  # Tasks validation logic
│   │   ├── implement.go              # Implementation validation
│   │   ├── parser.go                 # Markdown parsing utilities
│   │   └── retry.go                  # Retry state management
│   ├── claude/
│   │   ├── client.go                 # Claude CLI wrapper
│   │   ├── settings.go               # Settings generation
│   │   └── hooks.go                  # Hook management
│   ├── config/
│   │   ├── config.go                 # Config loading/saving
│   │   ├── defaults.go               # Default configuration
│   │   └── schema.go                 # Config struct definitions
│   ├── git/
│   │   ├── repo.go                   # Git operations
│   │   └── branch.go                 # Branch detection
│   └── install/
│       ├── installer.go              # Installation logic
│       └── paths.go                  # Path management
├── pkg/
│   └── speckit/
│       ├── types.go                  # Shared types
│       └── errors.go                 # Error definitions
├── scripts/                          # REFERENCE ONLY - delete after porting
│   ├── hooks/                        # Reference for Go implementation
│   ├── lib/                          # Logic to port to Go
│   └── *.sh                          # Delete once Go version complete
├── tests/                            # Keep bats tests as reference
├── go.mod
├── go.sum
├── Makefile                          # Build automation
├── PLAN.md                           # This file
└── README.md                         # Updated with Go installation

# After migration, final structure:
auto-claude-speckit/
├── cmd/autospec/main.go
├── internal/...                      # All Go code
├── testdata/                         # Test fixtures
├── .goreleaser.yml                   # Release automation
└── README.md
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)
**Goal**: Set up Go project structure and basic CLI

#### Tasks:
- [ ] Initialize Go module: `go mod init github.com/yourusername/autospec`
- [ ] Add dependencies:
  - `github.com/spf13/cobra` - CLI framework
  - `github.com/spf13/viper` - Configuration
  - `github.com/stretchr/testify` - Testing
  - `gopkg.in/yaml.v3` - YAML parsing (for markdown frontmatter)
- [ ] Create project structure (cmd/, internal/, pkg/)
- [ ] Implement root command with version, help
- [ ] Create Makefile with build targets
- [ ] Set up GitHub Actions for CI

#### Deliverables:
- Basic `autospec --version` and `autospec --help` working
- CI pipeline building binary for linux/darwin/windows

#### Files to Create:
```go
// cmd/autospec/main.go
package main

import (
    "github.com/yourusername/autospec/internal/cli"
    "os"
)

func main() {
    if err := cli.Execute(); err != nil {
        os.Exit(1)
    }
}
```

```go
// internal/cli/root.go
package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
    Use:   "autospec",
    Short: "Automated validation for Claude Code SpecKit workflows",
    Long:  `AutoSpec provides automated validation and workflow management...`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(workflowCmd)
    // ... other commands
}
```

---

### Phase 2: Core Validation Logic (Week 1-2)
**Goal**: Port bash validation library to Go

#### Tasks:
- [ ] Implement file existence validation
- [ ] Implement markdown parsing (extract tasks, phases)
- [ ] Port retry state management (use JSON files in ~/.autospec/state/)
- [ ] Implement task counting logic
- [ ] Port continuation prompt generation
- [ ] Create exit code constants matching bash version

#### Key Functions to Port:

From `scripts/lib/speckit-validation-lib.sh`:
- `validate_file_exists()` → `validator.FileExists(path string) bool`
- `count_unchecked_tasks()` → `validator.CountUncheckedTasks(file string) int`
- `get_retry_count()` → `retry.GetCount(spec, phase string) int`
- `increment_retry_count()` → `retry.Increment(spec, phase string) int`
- `generate_continuation_prompt()` → `validator.GenerateContinuationPrompt(...) string`
- `detect_current_spec()` → `git.DetectCurrentSpec() string`

#### Files to Create:

```go
// internal/validator/spec.go
package validator

import (
    "os"
    "path/filepath"
)

type SpecValidator struct {
    specsDir string
}

func NewSpecValidator(specsDir string) *SpecValidator {
    return &SpecValidator{specsDir: specsDir}
}

func (v *SpecValidator) ValidateSpecExists(specName, fileName string) error {
    // Find spec directory (may have number prefix like 002-)
    dirs, err := filepath.Glob(filepath.Join(v.specsDir, "*"+specName+"*"))
    if err != nil {
        return err
    }
    if len(dirs) == 0 {
        return ErrSpecNotFound
    }

    // Check if expected file exists
    filePath := filepath.Join(dirs[0], fileName)
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return ErrFileNotFound
    }

    return nil
}
```

```go
// internal/validator/tasks.go
package validator

import (
    "bufio"
    "os"
    "strings"
)

// CountUncheckedTasks counts incomplete tasks in tasks.md
func CountUncheckedTasks(filePath string) (int, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    count := 0
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        // Match unchecked task: - [ ] or * [ ]
        if strings.Contains(line, "- [ ]") || strings.Contains(line, "* [ ]") {
            count++
        }
    }

    return count, scanner.Err()
}

// ListIncompletePhases returns phases with unchecked tasks
func ListIncompletePhases(filePath string) ([]Phase, error) {
    // Parse markdown structure
    // Extract phases (## headings)
    // Count unchecked tasks per phase
    // Return phases with count > 0
}
```

```go
// internal/validator/retry.go
package validator

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type RetryState struct {
    Spec   string
    Phase  string
    Count  int
}

func GetRetryStateFile() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".autospec", "state", "retry.json")
}

func GetRetryCount(spec, phase string) int {
    // Load retry state from JSON
    // Return count for spec+phase
}

func IncrementRetryCount(spec, phase string) int {
    // Load, increment, save
}

func ResetRetryCount(spec, phase string) {
    // Remove entry from state
}
```

---

### Phase 3: CLI Commands (Week 2)
**Goal**: Implement all CLI commands

#### Tasks:
- [ ] `autospec init` - Create .autospec/config.json
- [ ] `autospec workflow <feature>` - Run full workflow
- [ ] `autospec specify <feature>` - Run specify phase
- [ ] `autospec plan` - Run plan phase
- [ ] `autospec tasks` - Run tasks phase
- [ ] `autospec implement` - Run implementation phase
- [ ] `autospec status` - Show implementation status
- [ ] `autospec config` - Show/edit configuration

#### Implementation Example:

```go
// internal/cli/workflow.go
package cli

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/yourusername/autospec/internal/claude"
    "github.com/yourusername/autospec/internal/config"
    "github.com/yourusername/autospec/internal/validator"
)

var workflowCmd = &cobra.Command{
    Use:   "workflow <feature-description>",
    Short: "Run complete SpecKit workflow with validation",
    Args:  cobra.MinimumNArgs(1),
    RunE:  runWorkflow,
}

func runWorkflow(cmd *cobra.Command, args []string) error {
    featureDesc := args[0]

    // Load config
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    // Create Claude client
    client := claude.NewClient(cfg)

    // Create validator
    v := validator.NewSpecValidator(cfg.SpecsDir)

    // Run workflow steps
    fmt.Println("Step 1/3: Creating specification...")
    if err := runSpecifyWithValidation(client, v, featureDesc); err != nil {
        return err
    }

    fmt.Println("Step 2/3: Creating implementation plan...")
    if err := runPlanWithValidation(client, v, featureDesc); err != nil {
        return err
    }

    fmt.Println("Step 3/3: Generating tasks...")
    if err := runTasksWithValidation(client, v, featureDesc); err != nil {
        return err
    }

    fmt.Println("✓ Workflow completed successfully!")
    return nil
}

func runSpecifyWithValidation(client *claude.Client, v *validator.SpecValidator, feature string) error {
    maxRetries := 3

    for attempt := 1; attempt <= maxRetries; attempt++ {
        // Execute Claude command
        if err := client.Execute(fmt.Sprintf("/speckit.specify %s", feature)); err != nil {
            return fmt.Errorf("execute claude: %w", err)
        }

        // Validate spec.md exists
        if err := v.ValidateSpecExists(feature, "spec.md"); err == nil {
            fmt.Println("✓ Validation: spec.md created successfully")
            return nil
        }

        if attempt < maxRetries {
            fmt.Printf("✗ Validation: spec.md missing (attempt %d/%d)\n", attempt, maxRetries)
            fmt.Println("Retrying...")
        }
    }

    return fmt.Errorf("retry limit exhausted")
}
```

---

### Phase 4: Claude Integration (Week 2-3)
**Goal**: Implement Claude CLI wrapper and hook management

#### Tasks:
- [ ] Implement Claude command execution
- [ ] Generate temporary settings.json files
- [ ] Create hook script templates
- [ ] Handle Claude output streaming
- [ ] Parse Claude responses
- [ ] Make Claude command configurable

#### Implementation:

```go
// internal/claude/client.go
package claude

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "github.com/yourusername/autospec/internal/config"
)

type Client struct {
    cfg      *config.Config
    settings string // Path to temp settings file
}

func NewClient(cfg *config.Config) *Client {
    return &Client{cfg: cfg}
}

func (c *Client) Execute(prompt string) error {
    // Check if custom command template is configured
    if c.cfg.CustomClaudeCmd != "" {
        return c.executeCustomCommand(prompt)
    }

    // Use standard command
    return c.executeStandardCommand(prompt)
}

func (c *Client) executeStandardCommand(prompt string) error {
    // Generate temporary settings file
    settingsFile, err := c.generateSettings()
    if err != nil {
        return err
    }
    defer os.Remove(settingsFile)

    // Build command args
    args := []string{"--settings", settingsFile}
    args = append(args, c.cfg.ClaudeArgs...)
    args = append(args, prompt)

    // Execute Claude
    cmd := exec.Command(c.cfg.ClaudeCmd, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // Set API key env var
    if c.cfg.UseAPIKey {
        cmd.Env = append(os.Environ(), "ANTHROPIC_API_KEY="+os.Getenv("ANTHROPIC_API_KEY"))
    } else {
        // Explicitly clear API key
        cmd.Env = filterEnv(os.Environ(), "ANTHROPIC_API_KEY")
    }

    return cmd.Run()
}

func (c *Client) executeCustomCommand(prompt string) error {
    // Custom command template, e.g.:
    // ANTHROPIC_API_KEY="" claude -p "{{PROMPT}}" --dangerously-skip-permissions --verbose --output-format stream-json | claude-clean

    // Replace {{PROMPT}} placeholder
    cmdStr := strings.ReplaceAll(c.cfg.CustomClaudeCmd, "{{PROMPT}}", prompt)

    // Check if command contains pipe
    if strings.Contains(cmdStr, "|") {
        return c.executeCustomPipeline(cmdStr)
    }

    // No pipe, execute directly
    return c.executeCustomDirect(cmdStr)
}

func (c *Client) executeCustomPipeline(cmdStr string) error {
    // Parse pipeline: "cmd1 | cmd2 | cmd3"
    parts := strings.Split(cmdStr, "|")

    var cmds []*exec.Cmd
    for i, part := range parts {
        part = strings.TrimSpace(part)

        // Parse environment variables at start
        // E.g., "ANTHROPIC_API_KEY="" claude -p ..."
        envVars, cleanPart := parseEnvVars(part)

        // Parse command and args
        args := parseCommandLine(cleanPart)
        if len(args) == 0 {
            return fmt.Errorf("empty command in pipeline")
        }

        cmd := exec.Command(args[0], args[1:]...)

        // Set environment variables
        cmd.Env = os.Environ()
        for k, v := range envVars {
            cmd.Env = setEnv(cmd.Env, k, v)
        }

        // First command: no stdin
        // Middle commands: stdin from previous
        // Last command: stdout to terminal
        if i > 0 {
            cmd.Stdin = cmds[i-1].Stdout
        }

        if i < len(parts)-1 {
            var err error
            cmd.Stdout, err = cmd.StdoutPipe()
            if err != nil {
                return err
            }
        } else {
            cmd.Stdout = os.Stdout
        }

        cmd.Stderr = os.Stderr

        cmds = append(cmds, cmd)
    }

    // Start all commands
    for _, cmd := range cmds {
        if err := cmd.Start(); err != nil {
            return err
        }
    }

    // Wait for all commands
    for _, cmd := range cmds {
        if err := cmd.Wait(); err != nil {
            return err
        }
    }

    return nil
}

func (c *Client) executeCustomDirect(cmdStr string) error {
    // Parse environment variables
    envVars, cleanCmd := parseEnvVars(cmdStr)

    // Parse command and args
    args := parseCommandLine(cleanCmd)
    if len(args) == 0 {
        return fmt.Errorf("empty command")
    }

    cmd := exec.Command(args[0], args[1:]...)

    // Set environment variables
    cmd.Env = os.Environ()
    for k, v := range envVars {
        cmd.Env = setEnv(cmd.Env, k, v)
    }

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

// parseEnvVars extracts environment variables from start of command
// E.g., 'ANTHROPIC_API_KEY="" FOO=bar command args' -> {ANTHROPIC_API_KEY: "", FOO: bar}, "command args"
func parseEnvVars(cmdStr string) (map[string]string, string) {
    envVars := make(map[string]string)
    parts := strings.Fields(cmdStr)

    i := 0
    for i < len(parts) {
        if strings.Contains(parts[i], "=") {
            kv := strings.SplitN(parts[i], "=", 2)
            key := kv[0]
            value := ""
            if len(kv) > 1 {
                value = strings.Trim(kv[1], `"'`)
            }
            envVars[key] = value
            i++
        } else {
            break
        }
    }

    remainingCmd := strings.Join(parts[i:], " ")
    return envVars, remainingCmd
}

// parseCommandLine splits command line respecting quotes
func parseCommandLine(cmdStr string) []string {
    var args []string
    var current strings.Builder
    inQuote := false
    quoteChar := rune(0)

    for _, r := range cmdStr {
        switch {
        case r == '"' || r == '\'':
            if !inQuote {
                inQuote = true
                quoteChar = r
            } else if r == quoteChar {
                inQuote = false
                quoteChar = 0
            } else {
                current.WriteRune(r)
            }
        case r == ' ' && !inQuote:
            if current.Len() > 0 {
                args = append(args, current.String())
                current.Reset()
            }
        default:
            current.WriteRune(r)
        }
    }

    if current.Len() > 0 {
        args = append(args, current.String())
    }

    return args
}

func filterEnv(env []string, key string) []string {
    var filtered []string
    prefix := key + "="
    for _, e := range env {
        if !strings.HasPrefix(e, prefix) {
            filtered = append(filtered, e)
        }
    }
    return filtered
}

func setEnv(env []string, key, value string) []string {
    filtered := filterEnv(env, key)
    return append(filtered, key+"="+value)
}

func (c *Client) generateSettings() (string, error) {
    // Create temp file
    tmpFile, err := ioutil.TempFile("", "autospec-settings-*.json")
    if err != nil {
        return "", err
    }
    defer tmpFile.Close()

    // Build settings structure
    settings := Settings{
        Hooks: Hooks{
            Stop: c.cfg.EnabledHooks,
        },
        Permissions: Permissions{
            Allow: []string{"Bash(*)", "Read(*)", "Write(*)"},
        },
    }

    // Write JSON
    data, err := json.MarshalIndent(settings, "", "  ")
    if err != nil {
        return "", err
    }

    if _, err := tmpFile.Write(data); err != nil {
        return "", err
    }

    return tmpFile.Name(), nil
}

type Settings struct {
    Hooks       Hooks       `json:"hooks"`
    Permissions Permissions `json:"permissions"`
}

type Hooks struct {
    Stop []string `json:"Stop"`
}

type Permissions struct {
    Allow []string `json:"allow"`
}
```

---

### Phase 5: Configuration System (Week 3)
**Goal**: Implement configuration management

#### Tasks:
- [ ] Define configuration schema
- [ ] Implement config loading (global + local)
- [ ] Implement config initialization
- [ ] Support environment variable overrides
- [ ] Validate configuration

#### Configuration Schema:

```go
// internal/config/schema.go
package config

type Config struct {
    // Claude command configuration
    // Simple mode (default for most users):
    ClaudeCmd      string   `json:"claude_cmd" default:"claude"`

    // Advanced mode (custom command template):
    // Use {{PROMPT}} placeholder for the prompt injection point
    // Example: "ANTHROPIC_API_KEY=\"\" claude -p \"{{PROMPT}}\" --dangerously-skip-permissions --verbose --output-format stream-json | claude-clean"
    CustomClaudeCmd string  `json:"custom_claude_cmd"`

    // If CustomClaudeCmd is set, it takes precedence
    // Otherwise, ClaudeCmd + ClaudeArgs are used

    ClaudeArgs     []string `json:"claude_args"`  // Default: ["-p"]
    UseAPIKey      bool     `json:"use_api_key"`  // Default: true

    // Workflow settings
    MaxRetries     int      `json:"max_retries" default:"3"`
    SpecsDir       string   `json:"specs_dir" default:"specs"`

    // Hook settings
    EnabledHooks   []string `json:"enabled_hooks"`

    // Output processing
    OutputProcessor string  `json:"output_processor"` // e.g., "claude-clean"
}

func Load() (*Config, error) {
    // Load from ~/.autospec/config.json (global)
    global := loadGlobal()

    // Load from .autospec/config.json (local repo)
    local := loadLocal()

    // Merge (local overrides global)
    cfg := merge(global, local)

    // Apply env var overrides
    applyEnvOverrides(cfg)

    return cfg, nil
}

func (c *Config) Save(path string) error {
    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}
```

```go
// internal/cli/init.go
package cli

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize autospec in current repository",
    RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
    // Check if .autospec/ exists
    if _, err := os.Stat(".autospec"); !os.IsNotExist(err) {
        return fmt.Errorf("already initialized")
    }

    // Create .autospec directory
    if err := os.Mkdir(".autospec", 0755); err != nil {
        return err
    }

    // Create default config
    cfg := config.Default()
    if err := cfg.Save(".autospec/config.json"); err != nil {
        return err
    }

    fmt.Println("✓ Initialized autospec configuration")
    fmt.Println("  Edit .autospec/config.json to customize")

    return nil
}
```

---

### Phase 6: Testing (Week 3-4)
**Goal**: Achieve feature parity with bash version

#### Tasks:
- [ ] Port existing bats tests to Go tests
- [ ] Test all validators match bash behavior
- [ ] Test CLI commands
- [ ] Integration tests for full workflows
- [ ] Test configuration loading
- [ ] Test retry logic

#### Test Structure:

```go
// internal/validator/tasks_test.go
package validator

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCountUncheckedTasks(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        expected int
    }{
        {
            name: "no tasks",
            content: `# Title
Some content`,
            expected: 0,
        },
        {
            name: "all checked",
            content: `- [x] Task 1
- [x] Task 2`,
            expected: 0,
        },
        {
            name: "mixed",
            content: `- [x] Task 1
- [ ] Task 2
- [ ] Task 3`,
            expected: 2,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create temp file with content
            tmpFile := createTempFile(t, tt.content)
            defer os.Remove(tmpFile)

            // Count tasks
            count, err := CountUncheckedTasks(tmpFile)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, count)
        })
    }
}
```

```go
// internal/validator/validator_test.go
package validator

import (
    "testing"
    "path/filepath"
)

func TestValidateSpecExists_Parity(t *testing.T) {
    // Use test fixtures from tests/fixtures/
    fixturesDir := filepath.Join("..", "..", "tests", "fixtures")

    v := NewSpecValidator(fixturesDir)

    // Test cases matching bats tests
    err := v.ValidateSpecExists("mock-spec", "spec.md")
    assert.NoError(t, err)

    err = v.ValidateSpecExists("mock-spec", "missing.md")
    assert.Error(t, err)
}
```

#### Integration Tests:

```go
// cmd/autospec/integration_test.go
package main

import (
    "os"
    "os/exec"
    "testing"
)

func TestWorkflowCommand_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Build binary
    build := exec.Command("go", "build", "-o", "autospec-test")
    if err := build.Run(); err != nil {
        t.Fatal(err)
    }
    defer os.Remove("autospec-test")

    // Run workflow command
    cmd := exec.Command("./autospec-test", "workflow", "test feature", "--dry-run")
    output, err := cmd.CombinedOutput()

    assert.NoError(t, err)
    assert.Contains(t, string(output), "Step 1/3")
}
```

---

### Phase 7: Build & Release Automation (Week 4)
**Goal**: Automated builds and distribution

#### Tasks:
- [ ] Create Makefile for common tasks
- [ ] Set up GoReleaser
- [ ] Configure GitHub Actions for releases
- [ ] Generate checksums for binaries
- [ ] Create install script for binary download

#### Makefile:

```makefile
# Makefile
.PHONY: build test clean install release

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/autospec cmd/autospec/main.go

test:
	go test -v ./...

test-integration:
	go test -v -tags=integration ./...

clean:
	rm -rf bin/ dist/

install: build
	cp bin/autospec ~/.local/bin/autospec
	chmod +x ~/.local/bin/autospec

release:
	goreleaser release --clean

# Cross-compilation
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/autospec-linux-amd64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/autospec-linux-arm64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/autospec-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/autospec-darwin-arm64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/autospec-windows-amd64.exe
```

#### GoReleaser Config:

```yaml
# .goreleaser.yml
project_name: autospec

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - id: autospec
    main: ./cmd/autospec
    binary: autospec
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

release:
  github:
    owner: yourusername
    name: autospec

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
```

#### GitHub Actions:

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test -v ./...

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### Phase 8: Documentation & Migration (Week 4)
**Goal**: Update documentation and migration guide

#### Tasks:
- [ ] Update README.md with Go installation
- [ ] Create MIGRATION.md guide
- [ ] Add architecture documentation
- [ ] Create contribution guide
- [ ] Record demo GIF/video

#### Updated README:

```markdown
# Auto Claude SpecKit

> Single-binary CLI tool for automated Claude Code SpecKit validation

## Installation

### Homebrew (macOS/Linux)
```bash
brew install yourusername/tap/autospec
```

### Go Install
```bash
go install github.com/yourusername/autospec@latest
```

### Binary Download
```bash
# Linux
curl -L https://github.com/yourusername/autospec/releases/latest/download/autospec-linux-amd64.tar.gz | tar xz
sudo mv autospec /usr/local/bin/

# macOS
curl -L https://github.com/yourusername/autospec/releases/latest/download/autospec-darwin-amd64.tar.gz | tar xz
sudo mv autospec /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/yourusername/autospec/releases/latest/download/autospec-windows-amd64.zip" -OutFile "autospec.zip"
Expand-Archive autospec.zip
Move-Item autospec\autospec.exe C:\Windows\System32\
```

## Quick Start

```bash
# Initialize in your repo
cd my-project
autospec init

# Run full workflow
autospec workflow "Add user authentication"

# Or run individual steps
autospec specify "Add OAuth support"
autospec plan
autospec tasks
autospec implement

# Check status
autospec status
```

## Configuration

### Simple Configuration (Most Users)

Edit `.autospec/config.json`:

```json
{
  "claude_cmd": "claude",
  "claude_args": ["-p"],
  "use_api_key": true,
  "max_retries": 3,
  "specs_dir": "specs",
  "enabled_hooks": ["stop-speckit-implement"]
}
```

### Advanced Configuration (Custom Claude Command)

For users with highly customized Claude setups:

```json
{
  "custom_claude_cmd": "ANTHROPIC_API_KEY=\"\" claude -p \"{{PROMPT}}\" --dangerously-skip-permissions --verbose --output-format stream-json | claude-clean",
  "max_retries": 3,
  "specs_dir": "specs",
  "enabled_hooks": ["stop-speckit-implement"]
}
```

The `{{PROMPT}}` placeholder will be replaced with the actual prompt. The custom command supports:
- Environment variable prefixes (e.g., `ANTHROPIC_API_KEY=""`)
- Pipes to external processors (e.g., `| claude-clean`)
- Complex flag combinations
- Quote escaping

## Commands

- `autospec init` - Initialize in current repo
- `autospec workflow <feature>` - Run complete workflow
- `autospec specify <feature>` - Create specification
- `autospec plan` - Create implementation plan
- `autospec tasks` - Generate tasks
- `autospec implement` - Run implementation
- `autospec status` - Check implementation status
- `autospec config` - Show configuration
- `autospec version` - Show version

## Development

```bash
# Clone repo
git clone https://github.com/yourusername/autospec
cd autospec

# Build
make build

# Test
make test

# Install locally
make install
```
```

---

## Technical Design Details

### Dependencies

**Build-time Dependencies** (embedded in binary):
```go
// go.mod
module github.com/yourusername/autospec

go 1.21

require (
    github.com/spf13/cobra v1.8.0          // CLI framework
    github.com/spf13/viper v1.18.2         // Configuration management
    github.com/go-git/go-git/v5 v5.11.0    // Pure Go git (no git binary needed)
    github.com/stretchr/testify v1.8.4     // Testing framework
    gopkg.in/yaml.v3 v3.0.1                // YAML parsing
)
```

**Runtime Dependencies**: NONE ✅

The binary is completely self-contained. Users don't need:
- bash, jq, grep, sed (all logic in Go)
- git binary (using go-git library)
- Any external tools except `claude` CLI itself

### Error Handling

Define custom error types:

```go
// pkg/speckit/errors.go
package speckit

import "errors"

var (
    ErrSpecNotFound      = errors.New("spec directory not found")
    ErrFileNotFound      = errors.New("expected file not found")
    ErrRetryExhausted    = errors.New("retry limit exhausted")
    ErrInvalidArgs       = errors.New("invalid arguments")
    ErrMissingDeps       = errors.New("missing dependencies")
    ErrClaudeNotFound    = errors.New("claude command not found")
    ErrConfigInvalid     = errors.New("configuration invalid")
)
```

### Logging

Use structured logging:

```go
// internal/logger/logger.go
package logger

import (
    "log"
    "os"
)

var (
    Info  = log.New(os.Stdout, "INFO:  ", 0)
    Error = log.New(os.Stderr, "ERROR: ", 0)
    Debug = log.New(os.Stdout, "DEBUG: ", 0)
)

func SetDebug(enabled bool) {
    if !enabled {
        Debug.SetOutput(io.Discard)
    }
}
```

### State Management

Store retry state in JSON:

```json
// ~/.autospec/state/retry.json
{
  "states": [
    {
      "spec": "user-auth",
      "phase": "specify",
      "count": 1,
      "last_attempt": "2024-01-20T10:30:00Z"
    }
  ]
}
```

### Git Integration

**Option 1: Shell out to git command** (Simpler, requires git installed)
```go
// internal/git/repo.go
package git

import (
    "os/exec"
    "strings"
)

func DetectCurrentSpec() (string, error) {
    // Get current branch
    cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    branch := strings.TrimSpace(string(output))

    // Extract spec name from branch
    // Example: feature/user-auth -> user-auth
    parts := strings.Split(branch, "/")
    if len(parts) > 1 {
        return parts[len(parts)-1], nil
    }

    return branch, nil
}

func GetRepoRoot() (string, error) {
    cmd := exec.Command("git", "rev-parse", "--show-toplevel")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}
```

**Option 2: Pure Go with go-git library** (Recommended - zero dependencies)
```go
// internal/git/repo.go
package git

import (
    "github.com/go-git/go-git/v5"
    "path/filepath"
)

func DetectCurrentSpec() (string, error) {
    repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
        DetectDotGit: true,
    })
    if err != nil {
        return "", err
    }

    head, err := repo.Head()
    if err != nil {
        return "", err
    }

    // Get branch name
    branch := head.Name().Short()

    // Extract spec name from branch
    // Example: feature/user-auth -> user-auth
    parts := strings.Split(branch, "/")
    if len(parts) > 1 {
        return parts[len(parts)-1], nil
    }

    return branch, nil
}

func GetRepoRoot() (string, error) {
    repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
        DetectDotGit: true,
    })
    if err != nil {
        return "", err
    }

    wt, err := repo.Worktree()
    if err != nil {
        return "", err
    }

    return wt.Filesystem.Root(), nil
}
```

**Decision**: Use go-git library for truly zero dependencies. Users won't need git installed.
```

---

## Migration Checklist

### Pre-Migration
- [ ] Audit all bash scripts for edge cases
- [ ] Document all bash behavior to preserve
- [ ] Create comprehensive test suite
- [ ] Set up test fixtures from existing bats tests

### During Migration
- [ ] Port validation library (highest priority)
- [ ] Implement CLI commands
- [ ] Port all tests
- [ ] Verify feature parity
- [ ] Performance benchmarking

### Post-Migration
- [ ] Update all documentation
- [ ] Create migration guide for existing users
- [ ] Deprecation notice for bash scripts
- [ ] Release v1.0.0
- [ ] Archive bash version in `legacy/` branch

### Validation Criteria
- [ ] All 60+ bash tests have Go equivalents
- [ ] All tests pass
- [ ] Binary size < 15MB
- [ ] Startup time < 50ms
- [ ] Works on Linux, macOS, Windows
- [ ] No runtime dependencies required
- [ ] Backward compatible with existing configs

---

## Release Strategy

### Version Numbering
- `v0.x.x` - Development/beta (Go version in progress)
- `v1.0.0` - First stable Go release
- `v1.x.x` - Maintenance and features

### Release Timeline
- **Week 4**: v0.1.0-beta (basic CLI working)
- **Week 6**: v0.5.0-beta (feature complete)
- **Week 8**: v1.0.0 (stable release)

### Release Process
1. Tag version: `git tag v1.0.0`
2. Push tag: `git push origin v1.0.0`
3. GitHub Actions runs tests
4. GoReleaser builds binaries
5. Creates GitHub release with artifacts
6. Update package managers (brew, etc)

---

## Future Enhancements

### Phase 9: Advanced Features (Post v1.0)
- [ ] Interactive mode with TUI (bubbletea)
- [ ] Plugin system for custom validators
- [ ] AI-powered spec suggestions
- [ ] Cloud sync for retry state
- [ ] Metrics and analytics
- [ ] `autospec watch` - continuous validation
- [ ] Integration with GitHub Actions
- [ ] VS Code extension

### Phase 10: Package Managers
- [ ] Homebrew formula
- [ ] AUR package (Arch Linux)
- [ ] Apt repository (Ubuntu/Debian)
- [ ] Chocolatey package (Windows)
- [ ] Snap package

### Phase 11: Enterprise Features
- [ ] Team collaboration
- [ ] Shared configurations
- [ ] Audit logging
- [ ] Policy enforcement
- [ ] SSO integration

---

## Risk Mitigation

### Risk: Logic parity issues
**Mitigation**: Maintain bash version until Go version tested in production for 2+ months

### Risk: Performance regression
**Mitigation**: Benchmark both versions, ensure Go version is faster or equal

### Risk: Breaking changes for users
**Mitigation**:
- Auto-migrate old configs
- Support both .autospec/config.json and old .claude/settings.json
- Clear migration documentation

### Risk: Platform compatibility issues
**Mitigation**:
- Test on Linux, macOS, Windows in CI
- Use cross-platform file path handling
- No shell-specific assumptions

### Risk: Maintenance burden
**Mitigation**:
- Comprehensive test coverage (>80%)
- Clear contribution guidelines
- Well-documented code
- Minimal dependencies

---

## Success Metrics

### Technical
- ✅ Binary works on Linux/macOS/Windows
- ✅ No runtime dependencies
- ✅ All 60+ tests passing
- ✅ Install time < 30 seconds
- ✅ Binary size < 15MB

### User Experience
- ✅ One-command installation
- ✅ Works anywhere without configuration
- ✅ Clear error messages
- ✅ `--help` provides useful info
- ✅ Fast execution (< 5s for workflows)

### Adoption
- 🎯 10+ GitHub stars in first month
- 🎯 5+ contributors
- 🎯 100+ downloads
- 🎯 Positive user feedback

---

## Conclusion

This plan transitions Auto Claude SpecKit from a bash-based tool to a **100% pure Go** binary while preserving all existing functionality.

### Key Changes from Bash Version

| Aspect | Bash Version | Go Version |
|--------|-------------|------------|
| **Distribution** | Clone repo, copy scripts | Single binary download |
| **Installation** | Manual path editing | `go install` or download |
| **Dependencies** | bash, jq, git, grep, sed | None (pure Go) |
| **Platform Support** | Linux/macOS only | Linux/macOS/Windows |
| **Maintenance** | Shell scripts | Type-safe Go code |
| **Testing** | bats framework | Native Go tests |
| **Claude Command** | Hardcoded custom command | Configurable with template support |
| **Performance** | ~0.22s validation | <0.1s validation (faster) |
| **Binary Size** | N/A | ~10-15MB |

### Pure Go Implementation Confirmed ✅

- **NO bash scripts embedded or executed**
- **NO shell dependencies at runtime**
- **ALL validation logic rewritten in Go**
- **Git operations using go-git library**
- **Markdown parsing in pure Go**
- **Retry state management in Go**
- **Claude command execution with native Go process handling**

Bash scripts remain in the repo **only as reference** during porting, then moved to `legacy/` branch.

### Phased Approach Benefits

1. **Low risk** - gradual migration with validation at each step
2. **High quality** - comprehensive testing matches bash behavior
3. **Great UX** - single binary, easy install, works anywhere
4. **Maintainability** - clean Go code, easy contributions
5. **Performance** - faster execution than bash
6. **Portability** - works on any platform

**Estimated Timeline**: 4 weeks for MVP (v1.0.0)

**Next Steps**:
1. ✅ Review and approve this plan
2. Initialize Go project (Phase 1)
3. Begin validation library port (Phase 2)
4. Implement custom Claude command handling (Phase 4)
5. Comprehensive testing (Phase 6)
6. Release v1.0.0 (Week 4)
