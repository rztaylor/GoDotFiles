package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCheck(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")

	// Mock HOME to control platform.ConfigDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("command requires init - fail when missing", func(t *testing.T) {
		// profile list should require init
		err := rootCmd.PersistentPreRunE(profileListCmd, []string{})
		if err == nil {
			t.Error("expected error for uninitialized repo, got nil")
		}
	})

	t.Run("command exempted from init - success when missing", func(t *testing.T) {
		// version should NOT require init
		err := rootCmd.PersistentPreRunE(versionCmd, []string{})
		if err != nil {
			t.Errorf("expected no error for version command, got %v", err)
		}

		// init should NOT require init
		err = rootCmd.PersistentPreRunE(initCmd, []string{})
		if err != nil {
			t.Errorf("expected no error for init command, got %v", err)
		}
	})

	t.Run("command requires init - success when present", func(t *testing.T) {
		// Configure git user for test repo creation
		configureGitUserGlobal(t, tmpDir)

		// Initialize the repo
		if err := createNewRepo(gdfDir); err != nil {
			t.Fatalf("failed to initialize repo: %v", err)
		}

		// profile list should now succeed
		err := rootCmd.PersistentPreRunE(profileListCmd, []string{})
		if err != nil {
			t.Errorf("expected no error for initialized repo, got %v", err)
		}
	})
}
