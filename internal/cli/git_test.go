package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/git"
)

func TestSaveCommand(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		setupFiles    bool
		wantError     bool
		errorContains string
	}{
		{
			name:       "save with custom message",
			message:    "Add test configuration",
			setupFiles: true,
			wantError:  false,
		},
		{
			name:       "save with default message",
			message:    "",
			setupFiles: true,
			wantError:  false,
		},
		{
			name:       "save when no changes",
			message:    "",
			setupFiles: false,
			wantError:  false, // Should not error, just inform
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			tmpDir := t.TempDir()
			homeDir := filepath.Join(tmpDir, "home")
			gdfDir := filepath.Join(homeDir, ".gdf")

			// Set HOME for test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", homeDir)
			defer os.Setenv("HOME", oldHome)

			// Create home directory
			if err := os.MkdirAll(homeDir, 0755); err != nil {
				t.Fatalf("creating home dir: %v", err)
			}

			// Configure git user
			configureGitUserGlobal(t, homeDir)

			// Initialize repo
			repo, err := git.Init(gdfDir)
			if err != nil {
				t.Fatalf("git.Init() error = %v", err)
			}

			// Create initial commit
			testFile := filepath.Join(gdfDir, "config.yaml")
			if err := os.WriteFile(testFile, []byte("initial: true\n"), 0644); err != nil {
				t.Fatalf("writing initial file: %v", err)
			}
			if err := repo.Add("."); err != nil {
				t.Fatalf("initial add: %v", err)
			}
			if err := repo.Commit("Initial commit"); err != nil {
				t.Fatalf("initial commit: %v", err)
			}

			// Setup test files if needed
			if tt.setupFiles {
				content := []byte("test: true\n")
				if err := os.WriteFile(testFile, content, 0644); err != nil {
					t.Fatalf("writing test file: %v", err)
				}
			}

			// Run save command
			args := []string{}
			if tt.message != "" {
				args = append(args, tt.message)
			}

			err = runSave(nil, args)

			if tt.wantError {
				if err == nil {
					t.Error("runSave() expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("runSave() error = %v, want error containing %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("runSave() unexpected error = %v", err)
				}
			}

			// Verify commit was created if there were changes
			if tt.setupFiles && !tt.wantError {
				hasChanges, err := repo.HasChanges()
				if err != nil {
					t.Fatalf("HasChanges() error = %v", err)
				}
				if hasChanges {
					t.Error("HasChanges() = true after save, want false")
				}
			}
		})
	}
}

func TestPushCommand(t *testing.T) {
	tests := []struct {
		name          string
		setupRemote   bool
		wantError     bool
		errorContains string
	}{
		{
			name:          "push without remote",
			setupRemote:   false,
			wantError:     true,
			errorContains: "remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			tmpDir := t.TempDir()
			homeDir := filepath.Join(tmpDir, "home")
			gdfDir := filepath.Join(homeDir, ".gdf")

			// Set HOME for test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", homeDir)
			defer os.Setenv("HOME", oldHome)

			// Create home directory
			if err := os.MkdirAll(homeDir, 0755); err != nil {
				t.Fatalf("creating home dir: %v", err)
			}

			// Configure git user
			configureGitUserGlobal(t, homeDir)

			// Initialize repo
			if _, err := git.Init(gdfDir); err != nil {
				t.Fatalf("git.Init() error = %v", err)
			}

			// Run push command
			err := runPush(nil, []string{})

			if tt.wantError {
				if err == nil {
					t.Error("runPush() expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("runPush() error = %v, want error containing %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("runPush() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestPullCommand(t *testing.T) {
	tests := []struct {
		name          string
		setupRemote   bool
		wantError     bool
		errorContains string
	}{
		{
			name:          "pull without remote",
			setupRemote:   false,
			wantError:     true,
			errorContains: "remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			tmpDir := t.TempDir()
			homeDir := filepath.Join(tmpDir, "home")
			gdfDir := filepath.Join(homeDir, ".gdf")

			// Set HOME for test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", homeDir)
			defer os.Setenv("HOME", oldHome)

			// Create home directory
			if err := os.MkdirAll(homeDir, 0755); err != nil {
				t.Fatalf("creating home dir: %v", err)
			}

			// Configure git user
			configureGitUserGlobal(t, homeDir)

			// Initialize repo
			if _, err := git.Init(gdfDir); err != nil {
				t.Fatalf("git.Init() error = %v", err)
			}

			// Run pull command
			err := runPull(nil, []string{})

			if tt.wantError {
				if err == nil {
					t.Error("runPull() expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("runPull() error = %v, want error containing %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("runPull() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestSyncCommand(t *testing.T) {
	tests := []struct {
		name          string
		setupRemote   bool
		wantError     bool
		errorContains string
	}{
		{
			name:          "sync without remote",
			setupRemote:   false,
			wantError:     true,
			errorContains: "remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			tmpDir := t.TempDir()
			homeDir := filepath.Join(tmpDir, "home")
			gdfDir := filepath.Join(homeDir, ".gdf")

			// Set HOME for test
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", homeDir)
			defer os.Setenv("HOME", oldHome)

			// Create home directory
			if err := os.MkdirAll(homeDir, 0755); err != nil {
				t.Fatalf("creating home dir: %v", err)
			}

			// Configure git user
			configureGitUserGlobal(t, homeDir)

			// Initialize repo
			if _, err := git.Init(gdfDir); err != nil {
				t.Fatalf("git.Init() error = %v", err)
			}

			// Run sync command
			err := runSync(nil, []string{})

			if tt.wantError {
				if err == nil {
					t.Error("runSync() expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("runSync() error = %v, want error containing %q", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("runSync() unexpected error = %v", err)
				}
			}
		})
	}
}
