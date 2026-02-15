package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/shell"
)

func TestInstallShellCompletion(t *testing.T) {
	t.Run("bash completion install", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		path, err := installShellCompletion(shell.Bash)
		if err != nil {
			t.Fatalf("installShellCompletion(bash) error = %v", err)
		}

		wantPath := filepath.Join(home, ".local", "share", "bash-completion", "completions", "gdf")
		if path != wantPath {
			t.Fatalf("path = %q, want %q", path, wantPath)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading completion file: %v", err)
		}
		if !strings.Contains(string(content), "__start_gdf") {
			t.Fatalf("bash completion output missing expected marker:\n%s", string(content))
		}
	})

	t.Run("zsh completion install", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		path, err := installShellCompletion(shell.Zsh)
		if err != nil {
			t.Fatalf("installShellCompletion(zsh) error = %v", err)
		}

		wantPath := filepath.Join(home, ".zfunc", "_gdf")
		if path != wantPath {
			t.Fatalf("path = %q, want %q", path, wantPath)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading completion file: %v", err)
		}
		if !strings.Contains(string(content), "#compdef gdf") {
			t.Fatalf("zsh completion output missing expected marker:\n%s", string(content))
		}
	})

	t.Run("unknown shell", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		if _, err := installShellCompletion(shell.Unknown); err == nil {
			t.Fatal("expected error for unknown shell, got nil")
		}
	})
}
