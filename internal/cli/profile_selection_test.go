package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/config"
)

func TestResolveProfileSelection(t *testing.T) {
	t.Run("uses explicit profile as-is", func(t *testing.T) {
		profile, err := resolveProfileSelection(t.TempDir(), "work")
		if err != nil {
			t.Fatalf("resolveProfileSelection() error = %v", err)
		}
		if profile != "work" {
			t.Fatalf("profile = %q, want work", profile)
		}
	})

	t.Run("errors when no profiles exist", func(t *testing.T) {
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

		_, err := resolveProfileSelection(gdfDir, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no profiles found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("auto-selects single profile", func(t *testing.T) {
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
		if err := writeProfile(gdfDir, "work", []string{"git"}); err != nil {
			t.Fatal(err)
		}

		profile, err := resolveProfileSelection(gdfDir, "")
		if err != nil {
			t.Fatalf("resolveProfileSelection() error = %v", err)
		}
		if profile != "work" {
			t.Fatalf("profile = %q, want work", profile)
		}
	})

	t.Run("prompts for multiple profiles", func(t *testing.T) {
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
		if err := writeProfile(gdfDir, "home", nil); err != nil {
			t.Fatal(err)
		}

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		oldStdin := os.Stdin
		os.Stdin = r
		defer func() { os.Stdin = oldStdin }()

		if _, err := w.Write([]byte("2\n")); err != nil {
			t.Fatal(err)
		}
		_ = w.Close()

		oldNonInteractive := globalNonInteractive
		globalNonInteractive = false
		defer func() { globalNonInteractive = oldNonInteractive }()

		profile, err := resolveProfileSelection(gdfDir, "")
		if err != nil {
			t.Fatalf("resolveProfileSelection() error = %v", err)
		}
		if profile != "work" {
			t.Fatalf("profile = %q, want work", profile)
		}
	})

	t.Run("cancel defaults to error", func(t *testing.T) {
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
		if err := writeProfile(gdfDir, "home", nil); err != nil {
			t.Fatal(err)
		}

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		oldStdin := os.Stdin
		os.Stdin = r
		defer func() { os.Stdin = oldStdin }()

		if _, err := w.Write([]byte("\n")); err != nil {
			t.Fatal(err)
		}
		_ = w.Close()

		_, err = resolveProfileSelection(gdfDir, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), profileSelectionPromptDefaultCancel) {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("multiple profiles in non-interactive mode returns exit code", func(t *testing.T) {
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
		if err := writeProfile(gdfDir, "home", nil); err != nil {
			t.Fatal(err)
		}

		oldNonInteractive := globalNonInteractive
		globalNonInteractive = true
		defer func() { globalNonInteractive = oldNonInteractive }()

		_, err := resolveProfileSelection(gdfDir, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if ExitCode(err) != exitCodeNonInteractiveStop {
			t.Fatalf("ExitCode(err) = %d, want %d", ExitCode(err), exitCodeNonInteractiveStop)
		}
		if !strings.Contains(err.Error(), "multiple profiles found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func writeProfile(gdfDir, name string, apps []string) error {
	profileDir := filepath.Join(gdfDir, "profiles", name)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return err
	}
	p := &config.Profile{Name: name, Apps: apps}
	return p.Save(filepath.Join(profileDir, "profile.yaml"))
}
