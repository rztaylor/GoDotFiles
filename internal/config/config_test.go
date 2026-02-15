package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		wantShell    string
		wantAliases  string
		wantDotfiles string
		wantErr      bool
	}{
		{
			name:         "empty config returns defaults",
			yaml:         "kind: Config/v1",
			wantShell:    "",
			wantAliases:  "last_wins",
			wantDotfiles: "error",
		},
		{
			name: "full config",
			yaml: `
kind: Config/v1
git:
  remote: git@github.com:user/dotfiles.git
  branch: main
shell: zsh
conflict_resolution:
  aliases: error
  dotfiles: backup_and_replace
package_manager:
  prefer:
    macos: brew
    linux: apt
security:
  confirm_scripts: true
  log_scripts: false
history:
  max_size_mb: 1024
ui:
  color: always
  color_section_headings: true
  highlight_key_values: false
`,
			wantShell:    "zsh",
			wantAliases:  "error",
			wantDotfiles: "backup_and_replace",
		},
		{
			name: "minimal config",
			yaml: `
kind: Config/v1
shell: bash
`,
			wantShell:    "bash",
			wantAliases:  "last_wins",
			wantDotfiles: "error",
		},
		{
			name: "invalid yaml",
			yaml: `kind: Config/v1
shell: [invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("writing temp file: %v", err)
			}

			cfg, err := LoadConfig(path)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}

			if cfg.Shell != tt.wantShell {
				t.Errorf("Shell = %q, want %q", cfg.Shell, tt.wantShell)
			}

			if tt.name == "full config" && (cfg.History == nil || cfg.History.MaxSizeMBDefault() != 1024) {
				t.Errorf("History.MaxSizeMBDefault() = %d, want 1024", cfg.History.MaxSizeMBDefault())
			}
			if tt.name == "full config" {
				if cfg.UI == nil {
					t.Fatal("UI config should be set for full config")
				}
				if got := cfg.UI.ColorDefault(); got != "always" {
					t.Fatalf("UI.ColorDefault() = %q, want always", got)
				}
				if !cfg.UI.ColorSectionHeadingsDefault() {
					t.Fatal("UI.ColorSectionHeadingsDefault() = false, want true")
				}
				if cfg.UI.HighlightKeyValuesDefault() {
					t.Fatal("UI.HighlightKeyValuesDefault() = true, want false")
				}
			}

			// Check conflict resolution defaults
			aliases := tt.wantAliases
			dotfiles := tt.wantDotfiles
			if cfg.ConflictResolution != nil {
				if got := cfg.ConflictResolution.AliasesDefault(); got != aliases {
					t.Errorf("AliasesDefault() = %q, want %q", got, aliases)
				}
				if got := cfg.ConflictResolution.DotfilesDefault(); got != dotfiles {
					t.Errorf("DotfilesDefault() = %q, want %q", got, dotfiles)
				}
			}
		})
	}
}

func TestLoadConfig_NonExistent(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil for missing file", err)
	}
	if cfg == nil {
		t.Error("LoadConfig() returned nil, want empty config")
	}
}

func TestConfig_Save(t *testing.T) {
	cfg := &Config{
		Shell: "zsh",
		ConflictResolution: &ConflictResolution{
			Aliases:  "last_wins",
			Dotfiles: "error",
		},
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")

	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if loaded.Shell != cfg.Shell {
		t.Errorf("Shell = %q, want %q", loaded.Shell, cfg.Shell)
	}

	if loaded.Kind != "Config/v1" {
		t.Errorf("Kind = %q, want Config/v1", loaded.Kind)
	}
}

func TestGitConfig_BranchDefault(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{"empty defaults to main", "", "main"},
		{"explicit branch", "develop", "develop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GitConfig{Branch: tt.branch}
			if got := g.BranchDefault(); got != tt.want {
				t.Errorf("BranchDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSecurityConfig_Defaults(t *testing.T) {
	// Nil values default to true
	s := &SecurityConfig{}
	if !s.ConfirmScriptsDefault() {
		t.Error("ConfirmScriptsDefault() = false, want true")
	}
	if !s.LogScriptsDefault() {
		t.Error("LogScriptsDefault() = false, want true")
	}

	// Explicit false
	f := false
	s2 := &SecurityConfig{ConfirmScripts: &f, LogScripts: &f}
	if s2.ConfirmScriptsDefault() {
		t.Error("ConfirmScriptsDefault() = true, want false")
	}
	if s2.LogScriptsDefault() {
		t.Error("LogScriptsDefault() = true, want false")
	}
}

func TestHistoryConfig_Defaults(t *testing.T) {
	t.Run("nil defaults to 512", func(t *testing.T) {
		var h *HistoryConfig
		if got := h.MaxSizeMBDefault(); got != 512 {
			t.Fatalf("MaxSizeMBDefault() = %d, want 512", got)
		}
	})

	t.Run("empty defaults to 512", func(t *testing.T) {
		h := &HistoryConfig{}
		if got := h.MaxSizeMBDefault(); got != 512 {
			t.Fatalf("MaxSizeMBDefault() = %d, want 512", got)
		}
	})

	t.Run("non-positive defaults to 512", func(t *testing.T) {
		zero := 0
		h := &HistoryConfig{MaxSizeMB: &zero}
		if got := h.MaxSizeMBDefault(); got != 512 {
			t.Fatalf("MaxSizeMBDefault() = %d, want 512", got)
		}
	})

	t.Run("positive value is used", func(t *testing.T) {
		v := 2048
		h := &HistoryConfig{MaxSizeMB: &v}
		if got := h.MaxSizeMBDefault(); got != 2048 {
			t.Fatalf("MaxSizeMBDefault() = %d, want 2048", got)
		}
	})
}

func TestShellIntegrationConfig_AutoReloadEnabledDefault(t *testing.T) {
	t.Run("nil shell integration defaults disabled", func(t *testing.T) {
		var s *ShellIntegrationConfig
		if s.AutoReloadEnabledDefault() {
			t.Fatal("AutoReloadEnabledDefault() = true, want false")
		}
	})

	t.Run("nil value defaults disabled", func(t *testing.T) {
		s := &ShellIntegrationConfig{}
		if s.AutoReloadEnabledDefault() {
			t.Fatal("AutoReloadEnabledDefault() = true, want false")
		}
	})

	t.Run("explicit enabled true", func(t *testing.T) {
		v := true
		s := &ShellIntegrationConfig{AutoReloadEnabled: &v}
		if !s.AutoReloadEnabledDefault() {
			t.Fatal("AutoReloadEnabledDefault() = false, want true")
		}
	})
}

func TestDefaultShell(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "bash", in: "bash", want: "bash"},
		{name: "zsh", in: "zsh", want: "zsh"},
		{name: "fish", in: "fish", want: "fish"},
		{name: "uppercase", in: "ZSH", want: "zsh"},
		{name: "unknown defaults to zsh", in: "tcsh", want: "zsh"},
		{name: "empty defaults to zsh", in: "", want: "zsh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultShell(tt.in); got != tt.want {
				t.Fatalf("DefaultShell(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDefaultConfigYAML_IncludesAllSections(t *testing.T) {
	content := DefaultConfigYAML("bash")
	required := []string{
		"kind: Config/v1",
		"git:",
		"remote: \"\"",
		"branch: main",
		"shell: bash",
		"conflict_resolution:",
		"aliases: last_wins",
		"dotfiles: error",
		"package_manager:",
		"prefer:",
		"macos: auto",
		"linux: auto",
		"wsl: auto",
		"security:",
		"confirm_scripts: true",
		"log_scripts: true",
		"history:",
		"max_size_mb: 512",
		"updates:",
		"disabled: false",
		"check_interval: 24h",
		"shell_integration:",
		"auto_reload_enabled: false",
		"ui:",
		"color: auto",
		"color_section_headings: true",
		"highlight_key_values: true",
	}

	for _, needle := range required {
		if !strings.Contains(content, needle) {
			t.Fatalf("DefaultConfigYAML missing %q:\n%s", needle, content)
		}
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nested", "config.yaml")

	if err := WriteDefaultConfig(path, "fish"); err != nil {
		t.Fatalf("WriteDefaultConfig() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written config: %v", err)
	}
	if !strings.Contains(string(data), "shell: fish") {
		t.Fatalf("expected shell default to be fish, got:\n%s", string(data))
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Kind != "Config/v1" {
		t.Fatalf("kind = %q, want Config/v1", cfg.Kind)
	}
}

func TestUIConfig_Defaults(t *testing.T) {
	t.Run("nil defaults", func(t *testing.T) {
		var u *UIConfig
		if got := u.ColorDefault(); got != "auto" {
			t.Fatalf("ColorDefault() = %q, want auto", got)
		}
		if !u.ColorSectionHeadingsDefault() {
			t.Fatal("ColorSectionHeadingsDefault() = false, want true")
		}
		if !u.HighlightKeyValuesDefault() {
			t.Fatal("HighlightKeyValuesDefault() = false, want true")
		}
	})

	t.Run("explicit values", func(t *testing.T) {
		section := false
		keys := false
		u := &UIConfig{
			Color:                "never",
			ColorSectionHeadings: &section,
			HighlightKeyValues:   &keys,
		}
		if got := u.ColorDefault(); got != "never" {
			t.Fatalf("ColorDefault() = %q, want never", got)
		}
		if u.ColorSectionHeadingsDefault() {
			t.Fatal("ColorSectionHeadingsDefault() = true, want false")
		}
		if u.HighlightKeyValuesDefault() {
			t.Fatal("HighlightKeyValuesDefault() = true, want false")
		}
	})
}
