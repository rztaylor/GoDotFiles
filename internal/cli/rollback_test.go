package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRollbackCommand(t *testing.T) {
	if rollbackCmd == nil {
		t.Fatal("rollbackCmd is nil")
	}
	if rollbackCmd.Use != "rollback" {
		t.Fatalf("rollbackCmd.Use = %s, want rollback", rollbackCmd.Use)
	}
}

func TestRunRollbackNoLogs(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	gdfDir := filepath.Join(tmpDir, ".gdf")
	if err := os.MkdirAll(gdfDir, 0755); err != nil {
		t.Fatal(err)
	}

	rollbackYes = true
	rollbackChooseSnapshot = false
	rollbackTarget = ""
	defer func() {
		rollbackYes = false
		rollbackChooseSnapshot = false
		rollbackTarget = ""
	}()

	if err := runRollback(nil, nil); err != nil {
		t.Fatalf("runRollback() error = %v", err)
	}
}
