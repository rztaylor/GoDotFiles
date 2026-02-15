package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAddProfileResolution(t *testing.T) {
	t.Run("adds app to sole profile when --profile is omitted", func(t *testing.T) {
		home := t.TempDir()
		gdfDir := filepath.Join(home, ".gdf")
		t.Setenv("HOME", home)
		configureGitUserGlobal(t, home)
		if err := createNewRepo(gdfDir); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(filepath.Join(gdfDir, "profiles")); err != nil {
			t.Fatal(err)
		}
		if err := writeProfile(gdfDir, "work", nil); err != nil {
			t.Fatal(err)
		}

		oldProfile := targetProfile
		oldNonInteractive := globalNonInteractive
		oldFromRecipe := fromRecipe
		targetProfile = ""
		globalNonInteractive = true
		fromRecipe = true
		defer func() {
			targetProfile = oldProfile
			globalNonInteractive = oldNonInteractive
			fromRecipe = oldFromRecipe
		}()

		if err := runAdd(nil, []string{"git"}); err != nil {
			t.Fatalf("runAdd() error = %v", err)
		}

		profilePath := filepath.Join(gdfDir, "profiles", "work", "profile.yaml")
		content, err := os.ReadFile(profilePath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "- git") {
			t.Fatalf("expected app in work profile, got:\n%s", string(content))
		}
	})

	t.Run("returns clear error when no profiles exist and --profile is omitted", func(t *testing.T) {
		home := t.TempDir()
		gdfDir := filepath.Join(home, ".gdf")
		t.Setenv("HOME", home)
		configureGitUserGlobal(t, home)
		if err := createNewRepo(gdfDir); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(filepath.Join(gdfDir, "profiles")); err != nil {
			t.Fatal(err)
		}

		oldProfile := targetProfile
		targetProfile = ""
		defer func() { targetProfile = oldProfile }()

		err := runAdd(nil, []string{"git"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no profiles found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
