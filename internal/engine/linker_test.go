package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestLinker_Link(t *testing.T) {
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
	sourceFile := filepath.Join(gdfDir, "dotfiles", "app", "config")
	if err := os.WriteFile(sourceFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		strategy string
		setup    func()             // setup target state
		check    func(string) error // check result
		wantErr  bool
	}{
		{
			name:     "new link",
			strategy: "error",
			setup:    func() {}, // no existing target
			check: func(target string) error {
				info, err := os.Lstat(target)
				if err != nil {
					return err
				}
				if info.Mode()&os.ModeSymlink == 0 {
					return prettify("not a symlink")
				}
				dest, _ := os.Readlink(target)
				if dest != sourceFile {
					return prettify("wrong destination: %s", dest)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:     "existing conflict error",
			strategy: "error",
			setup: func() {
				_ = os.WriteFile(filepath.Join(homeDir, ".config"), []byte("exists"), 0644)
			},
			wantErr: true,
		},
		{
			name:     "existing conflict replace",
			strategy: "replace",
			setup: func() {
				_ = os.WriteFile(filepath.Join(homeDir, ".config"), []byte("exists"), 0644)
			},
			check: func(target string) error {
				dest, _ := os.Readlink(target)
				if dest != sourceFile {
					return prettify("did not replace correctly")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:     "existing conflict backup",
			strategy: "backup_and_replace",
			setup: func() {
				_ = os.WriteFile(filepath.Join(homeDir, ".config"), []byte("original"), 0644)
			},
			check: func(target string) error {
				// Check link
				dest, _ := os.Readlink(target)
				if dest != sourceFile {
					return prettify("did not link")
				}
				// Check backup
				bak := target + ".gdf.bak"
				content, err := os.ReadFile(bak)
				if err != nil {
					return prettify("backup not created")
				}
				if string(content) != "original" {
					return prettify("backup content mismatch")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:     "idempotent link",
			strategy: "error",
			setup: func() {
				_ = os.Symlink(sourceFile, filepath.Join(homeDir, ".config"))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup target from previous runs
			target := filepath.Join(homeDir, ".config")
			_ = os.Remove(target)
			_ = os.Remove(target + ".gdf.bak")

			if tt.setup != nil {
				tt.setup()
			}

			l := NewLinker(tt.strategy)
			l.SetHistoryManager(NewHistoryManager(gdfDir, 512))
			dotfile := apps.Dotfile{
				Source: "app/config",
				Target: "~/.config",
			}

			err := l.Link(dotfile, gdfDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Link() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.check != nil {
				if err := tt.check(target); err != nil {
					t.Errorf("Check failed: %v", err)
				}
			}

		})
	}
}

func TestLinker_CapturesSnapshotOnReplace(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	homeDir := filepath.Join(tmpDir, "home")
	os.Setenv("HOME", homeDir)

	if err := os.MkdirAll(filepath.Join(gdfDir, "dotfiles", "app"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	sourceFile := filepath.Join(gdfDir, "dotfiles", "app", "config")
	if err := os.WriteFile(sourceFile, []byte("new-content"), 0644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(homeDir, ".config")
	if err := os.WriteFile(target, []byte("old-content"), 0644); err != nil {
		t.Fatal(err)
	}

	l := NewLinker("replace")
	l.SetHistoryManager(NewHistoryManager(gdfDir, 512))
	err := l.Link(apps.Dotfile{
		Source: "app/config",
		Target: "~/.config",
	}, gdfDir)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	s := l.ConsumeConflictSnapshot(target)
	if s == nil {
		t.Fatal("expected conflict snapshot, got nil")
	}
	snapData, err := os.ReadFile(s.Path)
	if err != nil {
		t.Fatalf("reading snapshot: %v", err)
	}
	if string(snapData) != "old-content" {
		t.Fatalf("snapshot content = %q, want old-content", string(snapData))
	}
}

func TestLinker_Unlink(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	target := filepath.Join(homeDir, ".link")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create symlink
	if err := os.Symlink("somewhere", target); err != nil {
		t.Fatal(err)
	}

	l := NewLinker("error")
	dotfile := apps.Dotfile{Target: "~/.link"}

	if err := l.Unlink(dotfile); err != nil {
		t.Errorf("Unlink() error = %v", err)
	}

	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Error("Symlink not removed")
	}
}

func TestLinker_UnlinkManaged(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	homeDir := filepath.Join(tmpDir, "home")
	target := filepath.Join(homeDir, ".link")
	source := filepath.Join(gdfDir, "dotfiles", "app", "config")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(filepath.Join(gdfDir, "dotfiles", "app"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("config"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}

	l := NewLinker("error")
	l.SetHistoryManager(NewHistoryManager(gdfDir, 512))

	snap, err := l.UnlinkManaged(apps.Dotfile{
		Source: "app/config",
		Target: "~/.link",
	}, gdfDir)
	if err != nil {
		t.Fatalf("UnlinkManaged() error = %v", err)
	}
	if snap == nil {
		t.Fatal("expected rollback snapshot for managed symlink unlink")
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatal("expected managed symlink to be removed")
	}
}

func TestLinker_UnlinkManaged_SkipsUnmanagedSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	homeDir := filepath.Join(tmpDir, "home")
	target := filepath.Join(homeDir, ".link")
	otherSource := filepath.Join(tmpDir, "outside")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(otherSource, []byte("outside"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(otherSource, target); err != nil {
		t.Fatal(err)
	}

	l := NewLinker("error")
	snap, err := l.UnlinkManaged(apps.Dotfile{
		Source: "app/config",
		Target: "~/.link",
	}, gdfDir)
	if err != nil {
		t.Fatalf("UnlinkManaged() error = %v", err)
	}
	if snap != nil {
		t.Fatal("expected no snapshot when unmanaged symlink is skipped")
	}
	if _, err := os.Lstat(target); err != nil {
		t.Fatalf("expected unmanaged symlink to remain: %v", err)
	}
}

func prettify(format string, args ...interface{}) error {
	return filepath.ErrBadPattern // just a dummy error type for simplicity in test helper
}
