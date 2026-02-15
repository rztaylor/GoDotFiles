package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/state"
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

	t.Run("move omitted from with multiple profiles in non-interactive mode", func(t *testing.T) {
		createProfile("source_def", []string{"appA"})
		createProfile("target_def", []string{})

		oldNonInteractive := globalNonInteractive
		globalNonInteractive = true
		defer func() { globalNonInteractive = oldNonInteractive }()

		moveFromProfile = ""
		moveToProfile = "target_def"

		err := runMove(nil, []string{"appA"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if ExitCode(err) != exitCodeNonInteractiveStop {
			t.Fatalf("ExitCode(err) = %d, want %d", ExitCode(err), exitCodeNonInteractiveStop)
		}
	})

	t.Run("move omitted to with multiple profiles in non-interactive mode", func(t *testing.T) {
		createProfile("source_def", []string{"appB"})
		createProfile("target_def", []string{})

		oldNonInteractive := globalNonInteractive
		globalNonInteractive = true
		defer func() { globalNonInteractive = oldNonInteractive }()

		moveFromProfile = "source_def"
		moveToProfile = ""

		err := runMove(nil, []string{"appB"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if ExitCode(err) != exitCodeNonInteractiveStop {
			t.Fatalf("ExitCode(err) = %d, want %d", ExitCode(err), exitCodeNonInteractiveStop)
		}
	})
}

func TestMoveAppsWithApplyUpdatesStateForBothProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	createProfile := func(name string, apps []string) {
		profileDir := filepath.Join(gdfDir, "profiles", name)
		if err := os.MkdirAll(profileDir, 0755); err != nil {
			t.Fatalf("failed to create profile dir: %v", err)
		}
		profile := &config.Profile{Name: name, Apps: apps}
		if err := profile.Save(filepath.Join(profileDir, "profile.yaml")); err != nil {
			t.Fatalf("failed to save profile: %v", err)
		}
	}
	createProfile("source", []string{"git"})
	createProfile("target", []string{})

	oldFrom := moveFromProfile
	oldTo := moveToProfile
	oldApply := moveApply
	oldYes := globalYes
	defer func() {
		moveFromProfile = oldFrom
		moveToProfile = oldTo
		moveApply = oldApply
		globalYes = oldYes
	}()

	moveFromProfile = "source"
	moveToProfile = "target"
	moveApply = true
	globalYes = true

	if err := runMove(nil, []string{"git"}); err != nil {
		t.Fatalf("runMove() error = %v", err)
	}

	st, err := state.LoadFromDir(gdfDir)
	if err != nil {
		t.Fatalf("loading state: %v", err)
	}
	if len(st.AppliedProfiles) != 2 {
		t.Fatalf("expected 2 applied profiles, got %d", len(st.AppliedProfiles))
	}
	if st.AppliedProfiles[0].Name != "source" || st.AppliedProfiles[1].Name != "target" {
		t.Fatalf("unexpected applied profiles order: %+v", st.AppliedProfiles)
	}
}

func TestMoveAppsWithApplyNonInteractiveRequiresYes(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	createProfile := func(name string, apps []string) {
		profileDir := filepath.Join(gdfDir, "profiles", name)
		if err := os.MkdirAll(profileDir, 0755); err != nil {
			t.Fatalf("failed to create profile dir: %v", err)
		}
		profile := &config.Profile{Name: name, Apps: apps}
		if err := profile.Save(filepath.Join(profileDir, "profile.yaml")); err != nil {
			t.Fatalf("failed to save profile: %v", err)
		}
	}
	createProfile("source", []string{"git"})
	createProfile("target", []string{})

	oldFrom := moveFromProfile
	oldTo := moveToProfile
	oldApply := moveApply
	oldYes := globalYes
	oldNonInteractive := globalNonInteractive
	defer func() {
		moveFromProfile = oldFrom
		moveToProfile = oldTo
		moveApply = oldApply
		globalYes = oldYes
		globalNonInteractive = oldNonInteractive
	}()

	moveFromProfile = "source"
	moveToProfile = "target"
	moveApply = true
	globalYes = false
	globalNonInteractive = true

	err := runMove(nil, []string{"git"})
	if err == nil {
		t.Fatal("expected non-interactive confirmation error")
	}
	if ExitCode(err) != exitCodeNonInteractiveStop {
		t.Fatalf("ExitCode(err) = %d, want %d", ExitCode(err), exitCodeNonInteractiveStop)
	}
}
