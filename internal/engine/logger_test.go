package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLogger(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	if err := os.MkdirAll(gdfDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("log operations", func(t *testing.T) {
		logger := NewLogger(false)

		logger.Log("link", "~/.gitconfig", map[string]string{
			"source": "git/config",
		})
		logger.Log("package_install", "git", map[string]string{
			"manager": "brew",
		})

		ops := logger.Operations()
		if len(ops) != 2 {
			t.Errorf("got %d operations, want 2", len(ops))
		}

		if ops[0].Type != "link" {
			t.Errorf("ops[0].Type = %s, want link", ops[0].Type)
		}
		if ops[0].Target != "~/.gitconfig" {
			t.Errorf("ops[0].Target = %s, want ~/.gitconfig", ops[0].Target)
		}
		if ops[0].Details["source"] != "git/config" {
			t.Errorf("ops[0].Details[source] = %s, want git/config", ops[0].Details["source"])
		}
	})

	t.Run("save operations", func(t *testing.T) {
		logger := NewLogger(false)
		logger.Log("link", "~/.vimrc", map[string]string{"source": "vim/vimrc"})

		logPath, err := logger.Save(gdfDir)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		if logPath == "" {
			t.Error("Save() returned empty path")
		}

		// Verify file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("log file was not created")
		}

		// Verify content
		data, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("reading log file: %v", err)
		}

		var ops []Operation
		if err := json.Unmarshal(data, &ops); err != nil {
			t.Fatalf("unmarshaling log: %v", err)
		}

		if len(ops) != 1 {
			t.Errorf("got %d operations in log, want 1", len(ops))
		}
		if ops[0].Type != "link" {
			t.Errorf("ops[0].Type = %s, want link", ops[0].Type)
		}
	})

	t.Run("dry run does not save", func(t *testing.T) {
		logger := NewLogger(true)
		logger.Log("link", "~/.bashrc", nil)

		logPath, err := logger.Save(gdfDir)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		if logPath != "" {
			t.Error("dry run should not save log file")
		}

		if !logger.IsDryRun() {
			t.Error("IsDryRun() = false, want true")
		}
	})

	t.Run("empty logger does not save", func(t *testing.T) {
		logger := NewLogger(false)

		logPath, err := logger.Save(gdfDir)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		if logPath != "" {
			t.Error("empty logger should not save log file")
		}
	})
}
