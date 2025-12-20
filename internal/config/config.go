// autospec - Spec-Driven Development Automation
// Author: Ariel Frischer
// Source: https://github.com/ariel-frischer/autospec

// Package config provides hierarchical configuration management for autospec using koanf.
// Configuration is loaded with priority: environment variables > project config (.autospec/config.yml)
// > user config (~/.config/autospec/config.yml) > defaults. It supports both YAML and legacy JSON
// formats, with migration utilities for transitioning from JSON to YAML.
package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ariel-frischer/autospec/internal/cliagent"
	"github.com/ariel-frischer/autospec/internal/notify"
	"github.com/ariel-frischer/autospec/internal/worktree"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// ConfigSource tracks where a configuration value came from
type ConfigSource string

const (
	SourceDefault ConfigSource = "default"
	SourceUser    ConfigSource = "user"
	SourceProject ConfigSource = "project"
	SourceEnv     ConfigSource = "env"
)

// Configuration represents the autospec CLI tool configuration
type Configuration struct {
	// AgentPreset selects a built-in agent by name (e.g., "claude", "gemini", "cline").
	// Takes precedence over legacy claude_cmd/claude_args fields.
	// Can be set via AUTOSPEC_AGENT_PRESET env var.
	AgentPreset string `koanf:"agent_preset"`

	// CustomAgent provides structured configuration for custom agents.
	// Takes precedence over agent_preset and all other agent configuration.
	// Example:
	//   custom_agent:
	//     command: "claude"
	//     args: ["-p", "--verbose", "{{PROMPT}}"]
	//     env:
	//       ANTHROPIC_API_KEY: ""
	//     post_processor: "cclean"
	CustomAgent *cliagent.CustomAgentConfig `koanf:"custom_agent"`

	// DEPRECATED: Use custom_agent instead. CustomAgentCmd defines a custom agent command string.
	CustomAgentCmd string `koanf:"custom_agent_cmd"`

	// DEPRECATED: Use agent_preset instead. ClaudeCmd specifies the CLI command to invoke.
	ClaudeCmd string `koanf:"claude_cmd"`
	// DEPRECATED: Use agent_preset instead. ClaudeArgs specifies additional CLI arguments.
	ClaudeArgs []string `koanf:"claude_args"`
	// DEPRECATED: Use custom_agent_cmd instead. CustomClaudeCmd specifies a custom command template.
	CustomClaudeCmd string `koanf:"custom_claude_cmd"`

	MaxRetries        int    `koanf:"max_retries"`
	SpecsDir          string `koanf:"specs_dir"`
	StateDir          string `koanf:"state_dir"`
	SkipPreflight     bool   `koanf:"skip_preflight"`
	Timeout           int    `koanf:"timeout"`
	SkipConfirmations bool   `koanf:"skip_confirmations"` // Skip confirmation prompts (can also be set via AUTOSPEC_YES env var)
	// ImplementMethod sets the default execution mode for the implement command.
	// Valid values: "single-session" (legacy), "phases" (default), "tasks"
	// Can be overridden by CLI flags (--phases, --tasks) or env var AUTOSPEC_IMPLEMENT_METHOD
	ImplementMethod string `koanf:"implement_method"`

	// Notifications configures notification preferences for command and stage completion.
	// Supports sound, visual, or both notification types across macOS, Linux, and Windows.
	// Environment variable support via AUTOSPEC_NOTIFICATIONS_* prefix.
	Notifications notify.NotificationConfig `koanf:"notifications"`

	// MaxHistoryEntries sets the maximum number of command history entries to retain.
	// Oldest entries are pruned when this limit is exceeded.
	// Default: 500. Can be set via AUTOSPEC_MAX_HISTORY_ENTRIES env var.
	MaxHistoryEntries int `koanf:"max_history_entries"`

	// ViewLimit sets the number of recent specs displayed by the view command.
	// Default: 5. Can be set via AUTOSPEC_VIEW_LIMIT env var.
	ViewLimit int `koanf:"view_limit"`

	// Worktree configures worktree management settings.
	// Used by the 'autospec worktree' command for creating and managing git worktrees.
	Worktree *worktree.WorktreeConfig `koanf:"worktree"`

	// DefaultAgents stores the list of agent names to pre-select in future init prompts.
	// Set during 'autospec init' when user selects agents for configuration.
	// Can be set via AUTOSPEC_DEFAULT_AGENTS env var (comma-separated).
	DefaultAgents []string `koanf:"default_agents,omitempty"`
}

// LoadOptions configures how configuration is loaded
type LoadOptions struct {
	// ProjectConfigPath overrides the project config path (default: .autospec/config.yml)
	ProjectConfigPath string
	// UserConfigPath overrides the user config path (default: ~/.config/autospec/config.yml)
	// Useful for testing to provide a mock user config
	UserConfigPath string
	// WarningWriter receives deprecation warnings (default: os.Stderr)
	WarningWriter io.Writer
	// SkipWarnings suppresses deprecation warnings
	SkipWarnings bool
}

// Load loads configuration from user, project, and environment sources.
// Priority: Environment variables > Project config > User config > Defaults
//
// New YAML config paths:
//   - User config: ~/.config/autospec/config.yml (XDG compliant)
//   - Project config: .autospec/config.yml
//
// Legacy JSON config paths (deprecated, triggers migration warning):
//   - User config: ~/.autospec/config.json
//   - Project config: .autospec/config.json
func Load(projectConfigPath string) (*Configuration, error) {
	return LoadWithOptions(LoadOptions{ProjectConfigPath: projectConfigPath})
}

// LoadWithOptions loads configuration with custom options
func LoadWithOptions(opts LoadOptions) (*Configuration, error) {
	k := koanf.New(".")
	warningWriter := getWarningWriter(opts.WarningWriter)

	loadDefaults(k)

	if err := loadUserConfig(k, opts.UserConfigPath, warningWriter, opts.SkipWarnings); err != nil {
		return nil, err
	}

	if err := loadProjectConfig(k, opts.ProjectConfigPath, warningWriter, opts.SkipWarnings); err != nil {
		return nil, err
	}

	if err := loadEnvironmentConfig(k); err != nil {
		return nil, err
	}

	return finalizeConfig(k)
}

// getWarningWriter returns the warning writer or defaults to stderr
func getWarningWriter(w io.Writer) io.Writer {
	if w == nil {
		return os.Stderr
	}
	return w
}

// loadDefaults applies default configuration values
func loadDefaults(k *koanf.Koanf) {
	defaults := GetDefaults()
	for key, value := range defaults {
		k.Set(key, value)
	}
}

// loadUserConfig loads user-level config (YAML preferred, legacy JSON supported).
// If customPath is provided, it uses that path exclusively (for testing).
// Otherwise: Priority: YAML (~/.config/autospec/config.yml) > JSON (~/.autospec/config.json).
// Warns if both exist (YAML used, JSON ignored) or if only legacy JSON exists.
func loadUserConfig(k *koanf.Koanf, customPath string, warningWriter io.Writer, skipWarnings bool) error {
	// If custom path provided, use it exclusively (for testing)
	if customPath != "" {
		if fileExists(customPath) {
			if err := loadYAMLConfig(k, customPath, "user"); err != nil {
				return fmt.Errorf("loading user YAML config: %w", err)
			}
		}
		return nil
	}

	userYAMLPath, _ := UserConfigPath()
	legacyUserPath, _ := LegacyUserConfigPath()

	userYAMLExists := fileExists(userYAMLPath)
	legacyUserExists := fileExists(legacyUserPath)

	if userYAMLExists {
		if err := loadYAMLConfig(k, userYAMLPath, "user"); err != nil {
			return fmt.Errorf("loading user YAML config: %w", err)
		}
		warnLegacyExists(warningWriter, legacyUserPath, userYAMLPath, legacyUserExists, skipWarnings, "--user")
	} else if legacyUserExists {
		if err := loadLegacyJSONConfig(k, legacyUserPath, "user", warningWriter, skipWarnings, "--user"); err != nil {
			return fmt.Errorf("loading legacy user JSON config: %w", err)
		}
	}
	return nil
}

// loadProjectConfig loads project-level config (YAML preferred, legacy JSON supported).
// Supports custom path override (for testing). Falls back to legacy JSON with warning.
// Same priority/warning logic as loadUserConfig.
func loadProjectConfig(k *koanf.Koanf, customPath string, warningWriter io.Writer, skipWarnings bool) error {
	projectYAMLPath := ProjectConfigPath()
	if customPath != "" {
		projectYAMLPath = customPath
	}
	legacyProjectPath := LegacyProjectConfigPath()

	projectYAMLExists := fileExists(projectYAMLPath)
	legacyProjectExists := fileExists(legacyProjectPath)

	if projectYAMLExists {
		if err := loadYAMLConfig(k, projectYAMLPath, "project"); err != nil {
			return fmt.Errorf("loading project YAML config: %w", err)
		}
		warnLegacyExists(warningWriter, legacyProjectPath, projectYAMLPath, legacyProjectExists, skipWarnings, "--project")
	} else if legacyProjectExists {
		if err := loadLegacyJSONConfig(k, legacyProjectPath, "project", warningWriter, skipWarnings, "--project"); err != nil {
			return fmt.Errorf("loading legacy project JSON config: %w", err)
		}
	}
	return nil
}

// loadYAMLConfig validates and loads a YAML config file
func loadYAMLConfig(k *koanf.Koanf, path, configType string) error {
	if err := ValidateYAMLSyntax(path); err != nil {
		return fmt.Errorf("validating YAML syntax for %s config: %w", configType, err)
	}
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return fmt.Errorf("failed to load %s config %s: %w", configType, path, err)
	}
	return nil
}

// loadLegacyJSONConfig loads legacy JSON and warns about migration
func loadLegacyJSONConfig(k *koanf.Koanf, path, configType string, warningWriter io.Writer, skipWarnings bool, migrateFlag string) error {
	if err := k.Load(file.Provider(path), json.Parser()); err != nil {
		return fmt.Errorf("failed to load legacy %s config %s: %w", configType, path, err)
	}
	if !skipWarnings {
		fmt.Fprintf(warningWriter, "Warning: Using deprecated JSON config at %s\n", path)
		fmt.Fprintf(warningWriter, "  Run 'autospec config migrate %s' to migrate to YAML format.\n\n", migrateFlag)
	}
	return nil
}

// warnLegacyExists warns if legacy JSON exists alongside new YAML
func warnLegacyExists(warningWriter io.Writer, legacyPath, yamlPath string, legacyExists, skipWarnings bool, migrateFlag string) {
	if legacyExists && !skipWarnings {
		fmt.Fprintf(warningWriter, "Warning: Legacy JSON config found at %s (ignored, using %s)\n", legacyPath, yamlPath)
		fmt.Fprintf(warningWriter, "  Run 'autospec config migrate %s' to remove the legacy file.\n\n", migrateFlag)
	}
}

// loadEnvironmentConfig loads environment variable overrides
func loadEnvironmentConfig(k *koanf.Koanf) error {
	if err := k.Load(env.Provider("AUTOSPEC_", ".", envTransform), nil); err != nil {
		return fmt.Errorf("failed to load environment config: %w", err)
	}
	return nil
}

// finalizeConfig unmarshals, validates, and applies final transformations
func finalizeConfig(k *koanf.Koanf) (*Configuration, error) {
	return finalizeConfigWithWarnings(k, os.Stderr, false)
}

// finalizeConfigWithWarnings unmarshals and optionally warns about deprecations
func finalizeConfigWithWarnings(k *koanf.Koanf, warningWriter io.Writer, skipWarnings bool) (*Configuration, error) {
	var cfg Configuration
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := ValidateConfigValues(&cfg, "config"); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	cfg.StateDir = expandHomePath(cfg.StateDir)
	cfg.SpecsDir = expandHomePath(cfg.SpecsDir)

	if os.Getenv("AUTOSPEC_YES") != "" {
		cfg.SkipConfirmations = true
	}

	// Emit deprecation warnings for legacy config fields
	if !skipWarnings {
		emitLegacyWarnings(&cfg, warningWriter)
	}

	return &cfg, nil
}

// emitLegacyWarnings writes deprecation warnings for legacy agent configuration fields
func emitLegacyWarnings(cfg *Configuration, w io.Writer) {
	// Only warn if legacy fields are in use and new fields are not set
	if cfg.AgentPreset != "" || cfg.CustomAgent.IsValid() {
		// New fields are in use, no need to warn
		return
	}

	if cfg.CustomAgentCmd != "" {
		fmt.Fprintf(w, "Warning: 'custom_agent_cmd' is deprecated. Use structured 'custom_agent' instead.\n")
		fmt.Fprintf(w, "  Replace:\n")
		fmt.Fprintf(w, "    custom_agent_cmd: %q\n", cfg.CustomAgentCmd)
		fmt.Fprintf(w, "  With:\n")
		fmt.Fprintf(w, "    custom_agent:\n")
		fmt.Fprintf(w, "      command: \"your-command\"\n")
		fmt.Fprintf(w, "      args: [\"-p\", \"{{PROMPT}}\"]\n")
		fmt.Fprintf(w, "      env:  # optional\n")
		fmt.Fprintf(w, "        VAR: \"value\"\n")
		fmt.Fprintf(w, "      post_processor: \"cclean\"  # optional\n\n")
	}

	if cfg.CustomClaudeCmd != "" {
		fmt.Fprintf(w, "Warning: 'custom_claude_cmd' is deprecated. Use structured 'custom_agent' instead.\n\n")
	}

	// Warn about claude_cmd/claude_args only if they differ from defaults
	defaults := GetDefaults()
	defaultCmd := defaults["claude_cmd"].(string)
	defaultArgs := defaults["claude_args"].([]string)

	cmdDiffers := cfg.ClaudeCmd != "" && cfg.ClaudeCmd != defaultCmd
	argsDiffer := len(cfg.ClaudeArgs) > 0 && !stringSliceEqual(cfg.ClaudeArgs, defaultArgs)

	if cmdDiffers || argsDiffer {
		fmt.Fprintf(w, "Warning: 'claude_cmd' and 'claude_args' are deprecated. Use 'agent_preset' or 'custom_agent'.\n")
		fmt.Fprintf(w, "  For Claude CLI: agent_preset: claude\n")
		fmt.Fprintf(w, "  For custom tool: Use structured 'custom_agent' config.\n\n")
	}
}

// stringSliceEqual compares two string slices for equality
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// fileExists returns true if the file exists and is readable
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// envTransform converts environment variable names to config keys
// Example: AUTOSPEC_MAX_RETRIES -> max_retries
func envTransform(s string) string {
	return strings.Replace(strings.ToLower(strings.TrimPrefix(s, "AUTOSPEC_")), "_", "_", -1)
}

// expandHomePath expands ~ to the user's home directory
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

// GetAgent returns a CLI agent based on configuration priority.
// Priority: custom_agent > custom_agent_cmd > agent_preset > legacy fields > default (claude).
// Returns error if the selected agent is invalid or not found in registry.
func (c *Configuration) GetAgent() (cliagent.Agent, error) {
	// Highest priority: structured custom_agent config
	if c.CustomAgent.IsValid() {
		return cliagent.NewCustomAgentFromConfig(*c.CustomAgent)
	}

	// Second priority: legacy custom_agent_cmd string (deprecated)
	if c.CustomAgentCmd != "" {
		// Convert legacy string to structured config
		cfg, err := parseLegacyCustomAgentCmd(c.CustomAgentCmd)
		if err != nil {
			return nil, fmt.Errorf("parsing custom_agent_cmd: %w", err)
		}
		return cliagent.NewCustomAgentFromConfig(cfg)
	}

	// Third priority: agent_preset (built-in agent by name)
	if c.AgentPreset != "" {
		agent := cliagent.Get(c.AgentPreset)
		if agent == nil {
			return nil, fmt.Errorf("unknown agent preset %q; available: %v", c.AgentPreset, cliagent.List())
		}
		return agent, nil
	}

	// Fourth priority: legacy custom_claude_cmd (deprecated)
	if c.CustomClaudeCmd != "" {
		cfg, err := parseLegacyCustomAgentCmd(c.CustomClaudeCmd)
		if err != nil {
			return nil, fmt.Errorf("parsing custom_claude_cmd: %w", err)
		}
		return cliagent.NewCustomAgentFromConfig(cfg)
	}

	// Fifth priority: legacy claude_cmd + claude_args (deprecated)
	// If claude_cmd is explicitly set to something other than default "claude",
	// build a custom agent config from it
	if c.ClaudeCmd != "" && c.ClaudeCmd != "claude" {
		cfg := buildLegacyConfig(c.ClaudeCmd, c.ClaudeArgs)
		return cliagent.NewCustomAgentFromConfig(cfg)
	}

	// Default: use claude agent from registry
	agent := cliagent.Get("claude")
	if agent == nil {
		return nil, fmt.Errorf("default agent 'claude' not registered")
	}
	return agent, nil
}

// parseLegacyCustomAgentCmd attempts to parse a legacy command string into structured config.
// This is best-effort and may not handle all shell syntax.
func parseLegacyCustomAgentCmd(cmdStr string) (cliagent.CustomAgentConfig, error) {
	// Simple parsing: split by whitespace, first token is command, rest are args
	// This won't handle pipes or env vars, but gives a helpful error message
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return cliagent.CustomAgentConfig{}, fmt.Errorf("empty command")
	}

	// Check for shell features that aren't supported
	if strings.Contains(cmdStr, "|") {
		return cliagent.CustomAgentConfig{}, fmt.Errorf(
			"pipes not supported in custom_agent_cmd string; use structured custom_agent config instead:\n" +
				"  custom_agent:\n" +
				"    command: \"your-command\"\n" +
				"    args: [...]\n" +
				"    post_processor: \"cclean\"")
	}

	// Check for env var prefix (VAR=value at start)
	first := parts[0]
	if idx := strings.Index(first, "="); idx > 0 {
		return cliagent.CustomAgentConfig{}, fmt.Errorf(
			"environment variables not supported in custom_agent_cmd string; use structured custom_agent config instead:\n" +
				"  custom_agent:\n" +
				"    command: \"your-command\"\n" +
				"    args: [...]\n" +
				"    env:\n" +
				"      VAR_NAME: \"value\"")
	}

	return cliagent.CustomAgentConfig{
		Command: parts[0],
		Args:    parts[1:],
	}, nil
}

// buildLegacyConfig constructs a CustomAgentConfig from legacy claude_cmd + claude_args.
// The -p flag is expected in claude_args and gets the prompt placeholder as its value.
func buildLegacyConfig(cmd string, args []string) cliagent.CustomAgentConfig {
	// Build args list, inserting {{PROMPT}} after -p flag
	var newArgs []string
	hasPromptFlag := false
	for _, arg := range args {
		newArgs = append(newArgs, arg)
		if arg == "-p" {
			newArgs = append(newArgs, "{{PROMPT}}")
			hasPromptFlag = true
		}
	}
	// If no -p flag was found, append prompt at end
	if !hasPromptFlag {
		newArgs = append(newArgs, "-p", "{{PROMPT}}")
	}
	return cliagent.CustomAgentConfig{
		Command: cmd,
		Args:    newArgs,
	}
}
