package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDecisionAudit_Save(t *testing.T) {
	gdfDir := t.TempDir()
	audit := newDecisionAudit("gdf app import", false)
	audit.Record("~/.aws/credentials", "sensitive-path", "track-as-secret")
	audit.Record("dotfiles/git/.gitconfig", "repo-path-exists", "skip")

	path, err := audit.Save(gdfDir)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if path == "" {
		t.Fatal("Save() path is empty")
	}
	if filepath.Base(filepath.Dir(path)) != ".operations" {
		t.Fatalf("expected .operations directory, got %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}
	var records []decisionRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("unmarshal decision log: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("decision record count = %d, want 2", len(records))
	}
	if records[0].Command != "gdf app import" {
		t.Fatalf("record command = %q, want gdf app import", records[0].Command)
	}
}

func TestDecisionAudit_DisabledOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		disabled bool
		record   bool
	}{
		{name: "disabled", disabled: true, record: true},
		{name: "empty", disabled: false, record: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audit := newDecisionAudit("gdf app track", tt.disabled)
			if tt.record {
				audit.Record("~/.zshrc", "target-conflict", "overwrite")
			}
			path, err := audit.Save(t.TempDir())
			if err != nil {
				t.Fatalf("Save() error = %v", err)
			}
			if path != "" {
				t.Fatalf("Save() path = %q, want empty", path)
			}
		})
	}
}
