package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/config"
)

func TestMoveApps(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	configureGitUserGlobal(t, tmpDir)

	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// Helper to create a profile
	createProfile := func(name string, apps []string) {
		profileDir := filepath.Join(gdfDir, "profiles", name)
		if err := os.MkdirAll(profileDir, 0755); err != nil {
			t.Fatalf("failed to create profile dir: %v", err)
		}
		profile := &config.Profile{
			Name: name,
			Apps: apps,
		}
		if err := profile.Save(filepath.Join(profileDir, "profile.yaml")); err != nil {
			t.Fatalf("failed to save profile: %v", err)
		}
	}

	t.Run("move single app", func(t *testing.T) {
		createProfile("source_single", []string{"app1", "app2"})
		createProfile("target_single", []string{"app3"})

		// Reset flags
		moveFromProfile = "source_single"
		moveToProfile = "target_single"

		err := runMove(nil, []string{"app1"})
		if err != nil {
			t.Errorf("runMove() error = %v", err)
		}

		// Verify source
		source, _ := config.LoadProfile(filepath.Join(gdfDir, "profiles", "source_single", "profile.yaml"))
		if len(source.Apps) != 1 || source.Apps[0] != "app2" {
			t.Errorf("source apps = %v, want [app2]", source.Apps)
		}

		// Verify target
		target, _ := config.LoadProfile(filepath.Join(gdfDir, "profiles", "target_single", "profile.yaml"))
		if len(target.Apps) != 2 || target.Apps[1] != "app1" {
			t.Errorf("target apps = %v, want [app3, app1]", target.Apps)
		}
	})

	t.Run("move wildcard apps", func(t *testing.T) {
		createProfile("source_wild", []string{"git", "git-lfs", "vim", "zsh"})
		createProfile("target_wild", []string{})

		moveFromProfile = "source_wild"
		moveToProfile = "target_wild"

		err := runMove(nil, []string{"git*"})
		if err != nil {
			t.Errorf("runMove() error = %v", err)
		}

		source, _ := config.LoadProfile(filepath.Join(gdfDir, "profiles", "source_wild", "profile.yaml"))
		if len(source.Apps) != 2 { // vim, zsh left
			t.Errorf("source apps count = %d, want 2 (vim, zsh)", len(source.Apps))
		}

		target, _ := config.LoadProfile(filepath.Join(gdfDir, "profiles", "target_wild", "profile.yaml"))
		if len(target.Apps) != 2 { // git, git-lfs moved
			t.Errorf("target apps count = %d, want 2 (git, git-lfs)", len(target.Apps))
		}
	})

	t.Run("move validation errors", func(t *testing.T) {
		// Both empty
		moveFromProfile = ""
		moveToProfile = ""
		err := runMove(nil, []string{"app"})
		if err == nil || !strings.Contains(err.Error(), "must specify at least one") {
			t.Errorf("expected error for missing flags, got %v", err)
		}

		// Same profile
		moveFromProfile = "same"
		moveToProfile = "same"
		err = runMove(nil, []string{"app"})
		if err == nil || !strings.Contains(err.Error(), "same") {
			t.Errorf("expected error for same profile, got %v", err)
		}
	})

	t.Run("move default inference from", func(t *testing.T) {
		// To specified, From empty -> From=default
		createProfile("default", []string{"appA"})
		createProfile("target_def", []string{})

		moveFromProfile = ""
		moveToProfile = "target_def"

		err := runMove(nil, []string{"appA"})
		if err != nil {
			t.Errorf("runMove() error = %v", err)
		}

		target, _ := config.LoadProfile(filepath.Join(gdfDir, "profiles", "target_def", "profile.yaml"))
		if len(target.Apps) != 1 || target.Apps[0] != "appA" {
			t.Errorf("appA not moved to target: %v", target.Apps)
		}
	})

	t.Run("move default inference to", func(t *testing.T) {
		// From specified, To empty -> To=default
		createProfile("default", []string{})
		createProfile("source_def", []string{"appB"})

		moveFromProfile = "source_def"
		moveToProfile = ""

		err := runMove(nil, []string{"appB"})
		if err != nil {
			t.Errorf("runMove() error = %v", err)
		}

		def, _ := config.LoadProfile(filepath.Join(gdfDir, "profiles", "default", "profile.yaml"))
		if len(def.Apps) != 1 || def.Apps[0] != "appB" {
			t.Errorf("appB not moved to default: %v", def.Apps)
		}
	})
}
