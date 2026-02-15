package cli

import (
	"os"
	"path/filepath"
	"strings"
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
		Shell: &apps.Shell{
			Init: []apps.InitSnippet{
				{
					Name:   "path",
					Common: `export PATH="$HOME/.testapp/bin:$PATH"`,
				},
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

	// Verify generated shell init contains app startup snippet
	generatedPath := filepath.Join(gdfDir, "generated", "init.sh")
	script, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("reading generated init script: %v", err)
	}
	if !strings.Contains(string(script), `export PATH="$HOME/.testapp/bin:$PATH"`) {
		t.Errorf("generated init missing app startup snippet:\n%s", string(script))
	}
}

func TestApplyRecursiveDependencies(t *testing.T) {
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

	// 1. Create dependency app (dep-app) with a dotfile we can check
	depAppPath := filepath.Join(gdfDir, "apps", "dep-app.yaml")

	// Create source file for dep-app
	depSourcePath := filepath.Join(gdfDir, "dotfiles", "dep-app", "config")
	if err := os.MkdirAll(filepath.Dir(depSourcePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(depSourcePath, []byte("dep-config"), 0644); err != nil {
		t.Fatal(err)
	}

	depBundle := &apps.Bundle{
		Name: "dep-app",
		Package: &apps.Package{
			Custom: &apps.CustomInstall{
				Script: "echo 'installing dep-app'",
			},
		},
		Dotfiles: []apps.Dotfile{
			{
				Source: "dep-app/config",
				Target: "~/.config/dep-app/config",
			},
		},
	}
	if err := depBundle.Save(depAppPath); err != nil {
		t.Fatal(err)
	}

	// 2. Create main app (main-app) depending on dep-app
	mainAppPath := filepath.Join(gdfDir, "apps", "main-app.yaml")
	mainBundle := &apps.Bundle{
		Name:         "main-app",
		Dependencies: []string{"dep-app"},
	}
	if err := mainBundle.Save(mainAppPath); err != nil {
		t.Fatal(err)
	}

	// 3. Add ONLY main-app to profile
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = []string{"main-app"} // dep-app is NOT listed
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	// 4. Run apply - should SUCCEED if recursive loading works.
	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply failed: %v", err)
	}

	// Verify side effect: check if dep-app's dotfile was linked
	targetPath := filepath.Join(homeDir, ".config", "dep-app", "config")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("dependency dotfile not found: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("dependency target is not a symlink")
	}
	dest, _ := os.Readlink(targetPath)
	if dest != depSourcePath {
		t.Errorf("wrong link destination for dependency: got %s, want %s", dest, depSourcePath)
	}
}
