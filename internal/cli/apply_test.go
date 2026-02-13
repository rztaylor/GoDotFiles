package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
)

func TestApply(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	// Mock environment
	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create repo structure
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	// Create a test app with dotfiles
	appPath := filepath.Join(gdfDir, "apps", "testapp.yaml")
	// Create source file
	sourcePath := filepath.Join(gdfDir, "dotfiles", "testapp", "config")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("config_content"), 0644); err != nil {
		t.Fatal(err)
	}

	bundle := &apps.Bundle{
		Name: "testapp",
		Dotfiles: []apps.Dotfile{
			{
				Source: "testapp/config",
				Target: "~/.config/testapp",
			},
		},
	}
	if err := bundle.Save(appPath); err != nil {
		t.Fatal(err)
	}

	// Add app to default profile
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = append(profile.Apps, "testapp")
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	// Run apply
	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply: %v", err)
	}

	// Verify symlink
	targetPath := filepath.Join(homeDir, ".config", "testapp")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("lstat target: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("target is not a symlink")
	}
	dest, _ := os.Readlink(targetPath)
	if dest != sourcePath {
		t.Errorf("wrong link destination: got %s, want %s", dest, sourcePath)
	}
}
