package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjector_InjectSourceLine(t *testing.T) {
	tests := []struct {
		name         string
		shellType    ShellType
		existingRC   string
		wantContains string
		wantBackup   bool
		wantErr      bool
	}{
		{
			name:         "inject into new bashrc",
			shellType:    Bash,
			existingRC:   "",
			wantContains: "~/.gdf/generated/init.sh",
			wantBackup:   false,
			wantErr:      false,
		},
		{
			name:         "inject into existing bashrc",
			shellType:    Bash,
			existingRC:   "export PATH=$PATH:~/bin\n",
			wantContains: "~/.gdf/generated/init.sh",
			wantBackup:   true,
			wantErr:      false,
		},
		{
			name:         "inject into zshrc",
			shellType:    Zsh,
			existingRC:   "",
			wantContains: "~/.gdf/generated/init.sh",
			wantBackup:   false,
			wantErr:      false,
		},
		{
			name:      "unknown shell type",
			shellType: Unknown,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp home directory
			tmpHome := t.TempDir()
			originalHome := os.Getenv("HOME")
			defer os.Setenv("HOME", originalHome)
			os.Setenv("HOME", tmpHome)

			// Create existing RC file if specified
			var rcPath string
			if tt.existingRC != "" {
				if tt.shellType == Bash {
					rcPath = filepath.Join(tmpHome, ".bashrc")
				} else if tt.shellType == Zsh {
					rcPath = filepath.Join(tmpHome, ".zshrc")
				}
				if rcPath != "" {
					err := os.WriteFile(rcPath, []byte(tt.existingRC), 0644)
					if err != nil {
						t.Fatalf("Failed to create test RC file: %v", err)
					}
				}
			}

			injector := NewInjector()
			err := injector.InjectSourceLine(tt.shellType)

			if (err != nil) != tt.wantErr {
				t.Errorf("InjectSourceLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Determine RC path
			rcPath = injector.getRCPath(tt.shellType)
			if rcPath == "" {
				t.Fatal("getRCPath returned empty string")
			}

			// Check RC file contains source line
			content, err := os.ReadFile(rcPath)
			if err != nil {
				t.Fatalf("Failed to read RC file: %v", err)
			}

			if !strings.Contains(string(content), tt.wantContains) {
				t.Errorf("RC file missing expected content %q. Got:\n%s", tt.wantContains, string(content))
			}

			// Check backup was created if expected
			backupPath := rcPath + ".gdf.backup"
			_, backupErr := os.Stat(backupPath)
			backupExists := backupErr == nil

			if backupExists != tt.wantBackup {
				t.Errorf("Backup file exists = %v, want %v", backupExists, tt.wantBackup)
			}
		})
	}
}

func TestInjector_DuplicateInjection(t *testing.T) {
	// Test that injecting twice doesn't duplicate the source line
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpHome)

	// Create .bashrc so getRCPath returns .bashrc (not .bash_profile fallback)
	rcPath := filepath.Join(tmpHome, ".bashrc")
	os.WriteFile(rcPath, []byte("# existing content\n"), 0644)

	injector := NewInjector()

	// Inject first time
	err := injector.InjectSourceLine(Bash)
	if err != nil {
		t.Fatalf("First injection failed: %v", err)
	}

	firstContent, _ := os.ReadFile(rcPath)

	// Inject second time
	err = injector.InjectSourceLine(Bash)
	if err != nil {
		t.Fatalf("Second injection failed: %v", err)
	}

	secondContent, _ := os.ReadFile(rcPath)

	// Content should be identical - no duplication
	if string(firstContent) != string(secondContent) {
		t.Errorf("Second injection modified file. First:\n%s\nSecond:\n%s", firstContent, secondContent)
	}

	// Debug: print what we got
	t.Logf("File content after second injection:\n%s", secondContent)

	// Count occurrences of the actual source line (not just the path which appears in comment too)
	count := strings.Count(string(secondContent), "[ -f ~/.gdf/generated/init.sh ] && source")
	if count != 1 {
		t.Errorf("Source line appears %d times, want 1.\nFull content:\n%s", count, secondContent)
	}
}

func TestInjector_GetRCPath(t *testing.T) {
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpHome)

	injector := NewInjector()

	tests := []struct {
		name      string
		shellType ShellType
		want      string
	}{
		{
			name:      "bash with bashrc",
			shellType: Bash,
			want:      filepath.Join(tmpHome, ".bashrc"),
		},
		{
			name:      "zsh",
			shellType: Zsh,
			want:      filepath.Join(tmpHome, ".zshrc"),
		},
		{
			name:      "unknown",
			shellType: Unknown,
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injector.getRCPath(tt.shellType)
			// For bash, create .bashrc to ensure it's preferred
			if tt.shellType == Bash {
				os.WriteFile(filepath.Join(tmpHome, ".bashrc"), []byte{}, 0644)
			}
			got = injector.getRCPath(tt.shellType)
			if got != tt.want {
				t.Errorf("getRCPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
