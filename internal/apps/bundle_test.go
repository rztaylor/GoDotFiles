package apps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantName string
		wantDeps []string
		wantErr  bool
	}{
		{
			name:     "minimal bundle",
			yaml:     `name: kubectl`,
			wantName: "kubectl",
			wantDeps: nil,
		},
		{
			name: "bundle with dependencies",
			yaml: `
name: kubectl-neat
description: Clean up kubectl output
dependencies:
  - kubectl
  - krew
`,
			wantName: "kubectl-neat",
			wantDeps: []string{"kubectl", "krew"},
		},
		{
			name: "package-less bundle",
			yaml: `
name: mac-preferences
description: macOS system preferences
hooks:
  apply:
    - run: defaults write com.apple.dock autohide -bool true
      when: os == 'macos'
`,
			wantName: "mac-preferences",
		},
		{
			name: "full bundle",
			yaml: `
name: git
description: Git version control
package:
  brew: git
  apt:
    name: git
dotfiles:
  - source: git/config
    target: ~/.gitconfig
shell:
  aliases:
    gs: git status
    gp: git push
  completions:
    bash: git completion bash
    zsh: git completion zsh
companions:
  - delta
  - gh
`,
			wantName: "git",
		},
		{
			name:    "invalid yaml",
			yaml:    `name: [invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("writing temp file: %v", err)
			}

			// Load bundle
			bundle, err := Load(path)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Check name
			if bundle.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", bundle.Name, tt.wantName)
			}

			// Check dependencies
			if len(bundle.Dependencies) != len(tt.wantDeps) {
				t.Errorf("Dependencies = %v, want %v", bundle.Dependencies, tt.wantDeps)
			}
		})
	}
}

func TestLoadAll(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"kubectl.yaml": "name: kubectl",
		"git.yaml":     "name: git",
		"README.md":    "# Not a YAML file",
	}
	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	// Load all bundles
	bundles, err := LoadAll(tmpDir)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	// Should only load .yaml files
	if len(bundles) != 2 {
		t.Errorf("got %d bundles, want 2", len(bundles))
	}
}

func TestBundle_Save(t *testing.T) {
	bundle := &Bundle{
		Name:         "test-app",
		Description:  "A test application",
		Dependencies: []string{"dep1", "dep2"},
		Shell: &Shell{
			Aliases: map[string]string{
				"ta": "test-app",
			},
		},
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-app.yaml")

	// Save bundle
	if err := bundle.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reload and verify
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Name != bundle.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, bundle.Name)
	}
	if len(loaded.Dependencies) != len(bundle.Dependencies) {
		t.Errorf("Dependencies = %v, want %v", loaded.Dependencies, bundle.Dependencies)
	}
}

func TestBundle_Validate(t *testing.T) {
	tests := []struct {
		name    string
		bundle  Bundle
		wantErr bool
	}{
		{
			name:   "valid minimal bundle",
			bundle: Bundle{Name: "kubectl"},
		},
		{
			name: "valid bundle with all fields",
			bundle: Bundle{
				Name:         "kubectl",
				Description:  "Kubernetes CLI",
				Dependencies: []string{"kubernetes"},
				Dotfiles: []Dotfile{
					{Source: "kube/config", Target: "~/.kube/config"},
				},
			},
		},
		{
			name:    "missing name",
			bundle:  Bundle{},
			wantErr: true,
		},
		{
			name:    "invalid name - uppercase",
			bundle:  Bundle{Name: "Kubectl"},
			wantErr: true,
		},
		{
			name:    "invalid name - starts with number",
			bundle:  Bundle{Name: "123app"},
			wantErr: true,
		},
		{
			name: "dotfile missing source",
			bundle: Bundle{
				Name: "test",
				Dotfiles: []Dotfile{
					{Target: "~/.config"},
				},
			},
			wantErr: true,
		},
		{
			name: "dotfile missing target",
			bundle: Bundle{
				Name: "test",
				Dotfiles: []Dotfile{
					{Source: "config"},
				},
			},
			wantErr: true,
		},
		{
			name: "plugin missing name",
			bundle: Bundle{
				Name: "test",
				Plugins: []Plugin{
					{Install: "krew install neat"},
				},
			},
			wantErr: true,
		},
		{
			name: "custom install missing script",
			bundle: Bundle{
				Name: "test",
				Package: &Package{
					Custom: &CustomInstall{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bundle.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v", err)
				}
			}
		})
	}
}

func TestTargetMap_GetTarget(t *testing.T) {
	tests := []struct {
		name      string
		targetMap TargetMap
		os        string
		want      string
	}{
		{
			name: "macos specific",
			targetMap: TargetMap{
				Default: "~/.config",
				Macos:   "~/Library/Preferences",
			},
			os:   "macos",
			want: "~/Library/Preferences",
		},
		{
			name: "linux fallback to default",
			targetMap: TargetMap{
				Default: "~/.config",
				Macos:   "~/Library/Preferences",
			},
			os:   "linux",
			want: "~/.config",
		},
		{
			name: "wsl specific",
			targetMap: TargetMap{
				Default: "~/.azure/config",
				Wsl:     "/mnt/c/Users/user/.azure/config",
			},
			os:   "wsl",
			want: "/mnt/c/Users/user/.azure/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.targetMap.GetTarget(tt.os)
			if got != tt.want {
				t.Errorf("GetTarget(%q) = %q, want %q", tt.os, got, tt.want)
			}
		})
	}
}

func TestCustomInstall_ConfirmDefault(t *testing.T) {
	tests := []struct {
		name    string
		confirm *bool
		want    bool
	}{
		{
			name:    "nil defaults to true",
			confirm: nil,
			want:    true,
		},
		{
			name:    "explicit true",
			confirm: boolPtr(true),
			want:    true,
		},
		{
			name:    "explicit false",
			confirm: boolPtr(false),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CustomInstall{
				Script:  "echo test",
				Confirm: tt.confirm,
			}
			if got := c.ConfirmDefault(); got != tt.want {
				t.Errorf("ConfirmDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
