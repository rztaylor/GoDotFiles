package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLatestOperationLog(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	logDir := filepath.Join(gdfDir, ".operations")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}

	older := filepath.Join(logDir, "20250101-000000.json")
	newer := filepath.Join(logDir, "20260101-000000.json")
	writeOps := func(path string, ops []Operation) {
		t.Helper()
		b, _ := json.Marshal(ops)
		if err := os.WriteFile(path, b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeOps(older, []Operation{{Type: "link", Target: "a"}})
	writeOps(newer, []Operation{{Type: "link", Target: "b"}})

	path, ops, err := LatestOperationLog(gdfDir)
	if err != nil {
		t.Fatalf("LatestOperationLog() error = %v", err)
	}
	if path != newer {
		t.Fatalf("LatestOperationLog() path = %s, want %s", path, newer)
	}
	if len(ops) != 1 || ops[0].Target != "b" {
		t.Fatalf("LatestOperationLog() ops = %#v", ops)
	}
}

func TestRollbackOperationsRestoreSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "home", ".zshrc")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/tmp/new", target); err != nil {
		t.Fatal(err)
	}

	snapshotPath := filepath.Join(tmpDir, "snap")
	if err := os.WriteFile(snapshotPath, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	ops := []Operation{
		{
			Type:   "link",
			Target: target,
			Details: map[string]string{
				"snapshot_path": snapshotPath,
				"snapshot_kind": "file",
				"snapshot_mode": "0644",
			},
			Timestamp: time.Now(),
		},
	}
	res := RollbackOperations(tmpDir, ops, nil)
	if len(res.Failed) > 0 {
		t.Fatalf("RollbackOperations() failed: %v", res.Failed)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading restored target: %v", err)
	}
	if string(content) != "old" {
		t.Fatalf("restored content = %q, want old", string(content))
	}
}

func TestFindSnapshotCandidates(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	logDir := filepath.Join(gdfDir, ".operations")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(tmpDir, "home", ".vimrc")

	now := time.Now().UTC()
	ops := []Operation{
		{
			Type:      "link",
			Target:    target,
			Timestamp: now.Add(-time.Hour),
			Details: map[string]string{
				"snapshot_path":        "/snap/a",
				"snapshot_kind":        "file",
				"snapshot_captured_at": now.Add(-time.Hour).Format(time.RFC3339Nano),
			},
		},
		{
			Type:      "link",
			Target:    target,
			Timestamp: now,
			Details: map[string]string{
				"snapshot_path":        "/snap/b",
				"snapshot_kind":        "file",
				"snapshot_captured_at": now.Format(time.RFC3339Nano),
			},
		},
	}
	b, _ := json.Marshal(ops)
	if err := os.WriteFile(filepath.Join(logDir, "20260101-000000.json"), b, 0644); err != nil {
		t.Fatal(err)
	}

	candidates, err := FindSnapshotCandidates(gdfDir, target)
	if err != nil {
		t.Fatalf("FindSnapshotCandidates() error = %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("FindSnapshotCandidates() len = %d, want 2", len(candidates))
	}
	if candidates[0].SnapshotPath != "/snap/b" {
		t.Fatalf("expected newest snapshot first, got %s", candidates[0].SnapshotPath)
	}
}
