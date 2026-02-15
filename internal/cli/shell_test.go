package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunShellCompletion(t *testing.T) {
	t.Run("bash completion", func(t *testing.T) {
		cmd := shellCompletionCmd
		cmd.SetOut(&bytes.Buffer{})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := runShellCompletion(cmd, []string{"bash"}); err != nil {
			t.Fatalf("runShellCompletion() error = %v", err)
		}
		if !strings.Contains(out.String(), "__start_gdf") {
			t.Fatalf("expected bash completion output, got:\n%s", out.String())
		}
	})

	t.Run("zsh completion", func(t *testing.T) {
		cmd := shellCompletionCmd
		cmd.SetOut(&bytes.Buffer{})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := runShellCompletion(cmd, []string{"zsh"}); err != nil {
			t.Fatalf("runShellCompletion() error = %v", err)
		}
		if !strings.Contains(out.String(), "#compdef gdf") {
			t.Fatalf("expected zsh completion output, got:\n%s", out.String())
		}
	})

	t.Run("unsupported shell", func(t *testing.T) {
		cmd := shellCompletionCmd
		cmd.SetOut(&bytes.Buffer{})

		err := runShellCompletion(cmd, []string{"fish"})
		if err == nil {
			t.Fatal("expected error for unsupported shell, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported shell") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
