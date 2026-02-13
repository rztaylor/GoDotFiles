package cli

import (
	"testing"
)

func TestRestoreCommand(t *testing.T) {
	if restoreCmd == nil {
		t.Fatal("restoreCmd is nil")
	}
	if restoreCmd.Use != "restore" {
		t.Errorf("expected use 'restore', got '%s'", restoreCmd.Use)
	}
}
