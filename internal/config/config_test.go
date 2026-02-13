package config

import (
	"os"
	"path/filepath"
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
