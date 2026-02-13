package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/schema"
)

func TestLoadProfile(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantName    string
		wantApps    int
		wantInclude int
		wantErr     bool
	}{
		{
			name: "minimal profile",
			yaml: `
kind: Profile/v1
name: base
`,
			wantName: "base",
		},
		{
			name: "profile with apps",
			yaml: `
kind: Profile/v1
name: sre
description: SRE toolkit
apps:
  - kubectl
  - terraform
  - aws-cli
`,
			wantName: "sre",
			wantApps: 3,
		},
		{
			name: "profile with includes",
			yaml: `
kind: Profile/v1
name: work
includes:
  - base
  - sre
apps:
  - slack
`,
			wantName:    "work",
			wantInclude: 2,
			wantApps:    1,
		},
		{
			name: "profile with conditions",
			yaml: `
kind: Profile/v1
name: dev
apps:
  - git
conditions:
  - if: os == 'macos'
    include_apps:
      - iterm2
    exclude_apps:
      - gnome-terminal
`,
			wantName: "dev",
			wantApps: 1,
		},
		{
			name: "invalid yaml",
			yaml: `kind: Profile/v1
name: [invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "profile.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("writing temp file: %v", err)
			}

			profile, err := LoadProfile(path)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadProfile() error = %v", err)
			}

			if profile.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", profile.Name, tt.wantName)
			}
			if len(profile.Apps) != tt.wantApps {
				t.Errorf("Apps count = %d, want %d", len(profile.Apps), tt.wantApps)
			}
			if len(profile.Includes) != tt.wantInclude {
				t.Errorf("Includes count = %d, want %d", len(profile.Includes), tt.wantInclude)
			}
		})
	}
}

func TestLoadAllProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create profile directories
	profiles := map[string]string{
		"base": `kind: Profile/v1
name: base`,
		"sre": `kind: Profile/v1
name: sre`,
	}

	for name, content := range profiles {
		dir := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("creating dir: %v", err)
		}
		path := filepath.Join(dir, "profile.yaml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("writing profile: %v", err)
		}
	}

	// Create a non-profile directory (no profile.yaml)
	notProfile := filepath.Join(tmpDir, "not-a-profile")
	if err := os.MkdirAll(notProfile, 0755); err != nil {
		t.Fatalf("creating dir: %v", err)
	}

	loaded, err := LoadAllProfiles(tmpDir)
	if err != nil {
		t.Fatalf("LoadAllProfiles() error = %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("got %d profiles, want 2", len(loaded))
	}
}

func TestProfile_Validate(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr bool
	}{
		{
			name:    "valid profile",
			profile: Profile{TypeMeta: schema.TypeMeta{Kind: "Profile/v1"}, Name: "base"},
		},
		{
			name:    "missing name",
			profile: Profile{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
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

func TestProfile_Save(t *testing.T) {
	profile := &Profile{
		Name:        "test",
		Description: "Test profile",
		Apps:        []string{"app1", "app2"},
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profile.yaml")

	if err := profile.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error = %v", err)
	}

	if loaded.Name != profile.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, profile.Name)
	}
	if len(loaded.Apps) != len(profile.Apps) {
		t.Errorf("Apps = %v, want %v", loaded.Apps, profile.Apps)
	}

	if loaded.Kind != "Profile/v1" {
		t.Errorf("Kind = %q, want Profile/v1", loaded.Kind)
	}
}
