package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestLinker_Restore(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	homeDir := filepath.Join(tmpDir, "home")

	// Setup dirs
	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(filepath.Join(gdfDir, "dotfiles", "app"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create source file
	sourcePath := filepath.Join(gdfDir, "dotfiles", "app", "config")
	content := []byte("original content")
	if err := os.WriteFile(sourcePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		setup   func()
		check   func(string) error
		wantErr bool
	}{
		{
			name: "restore valid symlink",
			setup: func() {
				// Create symlink from home to source
				target := filepath.Join(homeDir, ".config")
				_ = os.Symlink(sourcePath, target)
			},
			check: func(target string) error {
				// Should be a regular file now
				info, err := os.Lstat(target)
				if err != nil {
					return err
				}
				if info.Mode()&os.ModeSymlink != 0 {
					return prettify("still a symlink")
				}
				// Content should match
				got, _ := os.ReadFile(target)
				if string(got) != "original content" {
					return prettify("content mismatch")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "ignore unrelated file",
			setup: func() {
				target := filepath.Join(homeDir, ".config")
				_ = os.WriteFile(target, []byte("different"), 0644)
			},
			check: func(target string) error {
				got, _ := os.ReadFile(target)
				if string(got) != "different" {
					return prettify("file was modified")
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := filepath.Join(homeDir, ".config")
			_ = os.Remove(target) // Cleanup

			if tt.setup != nil {
				tt.setup()
			}

			l := NewLinker("error")
			dotfile := apps.Dotfile{
				Source: "app/config",
				Target: "~/.config",
			}

			err := l.Restore(dotfile, gdfDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Restore() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.check != nil {
				if err := tt.check(target); err != nil {
					t.Errorf("Check failed: %v", err)
				}
			}
		})
	}
}
