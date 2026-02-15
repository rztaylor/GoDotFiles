package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultShell normalizes the provided shell and falls back to zsh.
func DefaultShell(shell string) string {
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "bash", "zsh", "fish":
		return strings.ToLower(strings.TrimSpace(shell))
	default:
		return "zsh"
	}
}

// DefaultConfigYAML returns a fully-expanded default config.yaml template.
func DefaultConfigYAML(shell string) string {
	return fmt.Sprintf(`# GDF Configuration
kind: Config/v1

git:
  remote: ""
  branch: main

shell: %s

conflict_resolution:
  aliases: last_wins
  dotfiles: error

package_manager:
  prefer:
    macos: auto
    linux: auto
    wsl: auto

security:
  confirm_scripts: true
  log_scripts: true

history:
  max_size_mb: 512

updates:
  disabled: false
  check_interval: 24h

shell_integration:
  auto_reload_enabled: false

ui:
  color: auto
  color_section_headings: true
  highlight_key_values: true
`, DefaultShell(shell))
}

// WriteDefaultConfig writes a fully-expanded default config.yaml.
func WriteDefaultConfig(path, shell string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(DefaultConfigYAML(shell)), 0644); err != nil {
		return fmt.Errorf("writing default config: %w", err)
	}
	return nil
}
