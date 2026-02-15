package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/rztaylor/GoDotFiles/internal/schema"
)

// Config represents the global GDF configuration (~/.gdf/config.yaml).
type Config struct {
	schema.TypeMeta `yaml:",inline"`

	// Git contains repository settings.
	Git *GitConfig `yaml:"git,omitempty"`

	// Shell is the default shell (zsh, bash, fish).
	Shell string `yaml:"shell,omitempty"`

	// ConflictResolution defines how to handle conflicts.
	ConflictResolution *ConflictResolution `yaml:"conflict_resolution,omitempty"`

	// PackageManager contains package manager preferences.
	PackageManager *PackageManagerConfig `yaml:"package_manager,omitempty"`

	// Security contains security settings.
	Security *SecurityConfig `yaml:"security,omitempty"`

	// History contains snapshot retention settings.
	History *HistoryConfig `yaml:"history,omitempty"`

	// Updates contains auto-update settings.
	Updates *UpdatesConfig `yaml:"updates,omitempty"`
}

// UpdatesConfig holds auto-update settings.
type UpdatesConfig struct {
	// Disabled disables auto-update checks.
	Disabled bool `yaml:"disabled,omitempty"`

	// CheckInterval is the interval between checks (default: 24h).
	CheckInterval *time.Duration `yaml:"check_interval,omitempty"`
}

// GitConfig holds git repository settings.
type GitConfig struct {
	// Remote is the git remote URL.
	Remote string `yaml:"remote,omitempty"`

	// Branch is the default branch (default: main).
	Branch string `yaml:"branch,omitempty"`
}

// BranchDefault returns the effective branch (default: main).
func (g *GitConfig) BranchDefault() string {
	if g.Branch == "" {
		return "main"
	}
	return g.Branch
}

// ConflictResolution defines conflict handling strategies.
type ConflictResolution struct {
	// Aliases strategy: last_wins, error, or prompt (default: last_wins).
	Aliases string `yaml:"aliases,omitempty"`

	// Dotfiles strategy: error, backup_and_replace, or prompt (default: error).
	Dotfiles string `yaml:"dotfiles,omitempty"`
}

// AliasesDefault returns the effective aliases strategy.
func (c *ConflictResolution) AliasesDefault() string {
	if c.Aliases == "" {
		return "last_wins"
	}
	return c.Aliases
}

// DotfilesDefault returns the effective dotfiles strategy.
func (c *ConflictResolution) DotfilesDefault() string {
	if c.Dotfiles == "" {
		return "error"
	}
	return c.Dotfiles
}

// PackageManagerConfig defines package manager preferences.
type PackageManagerConfig struct {
	// Prefer specifies which package manager to prefer per platform.
	Prefer *PackageManagerPrefer `yaml:"prefer,omitempty"`
}

// PackageManagerPrefer maps platforms to preferred package managers.
type PackageManagerPrefer struct {
	Macos string `yaml:"macos,omitempty"`
	Linux string `yaml:"linux,omitempty"`
	Wsl   string `yaml:"wsl,omitempty"`
}

// SecurityConfig holds security settings.
type SecurityConfig struct {
	// ConfirmScripts requires confirmation for custom scripts (default: true).
	ConfirmScripts *bool `yaml:"confirm_scripts,omitempty"`

	// LogScripts enables logging of script executions (default: true).
	LogScripts *bool `yaml:"log_scripts,omitempty"`
}

// HistoryConfig controls file snapshot retention.
type HistoryConfig struct {
	// MaxSizeMB is the maximum disk usage for ~/.gdf/.history in MB (default: 512).
	MaxSizeMB *int `yaml:"max_size_mb,omitempty"`
}

// ConfirmScriptsDefault returns the effective confirm_scripts value.
func (s *SecurityConfig) ConfirmScriptsDefault() bool {
	if s.ConfirmScripts == nil {
		return true
	}
	return *s.ConfirmScripts
}

// LogScriptsDefault returns the effective log_scripts value.
func (s *SecurityConfig) LogScriptsDefault() bool {
	if s.LogScripts == nil {
		return true
	}
	return *s.LogScripts
}

// MaxSizeMBDefault returns the effective history.max_size_mb value.
func (h *HistoryConfig) MaxSizeMBDefault() int {
	if h == nil || h.MaxSizeMB == nil || *h.MaxSizeMB <= 0 {
		return 512
	}
	return *h.MaxSizeMB
}

// LoadConfig reads the global config from path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}

	if err := cfg.ValidateKind("Config"); err != nil {
		return nil, fmt.Errorf("validating config version: %w", err)
	}

	return &cfg, nil
}

// LoadConfigFromDir loads config.yaml from a directory.
func LoadConfigFromDir(dir string) (*Config, error) {
	return LoadConfig(filepath.Join(dir, "config.yaml"))
}

// Save writes the config to a file.
func (c *Config) Save(path string) error {
	c.Kind = "Config/v1"
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
