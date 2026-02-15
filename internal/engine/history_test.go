package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryManagerCaptureFile(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	target := filepath.Join(tmpDir, "home", ".zshrc")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	h := NewHistoryManager(gdfDir, 512)
	s, err := h.Capture(target)
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	if s == nil {
		t.Fatal("Capture() returned nil snapshot")
	}
	if s.Kind != "file" {
		t.Fatalf("snapshot kind = %s, want file", s.Kind)
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		t.Fatalf("reading snapshot: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("snapshot data = %q, want hello", string(data))
	}
}

func TestHistoryManagerCaptureSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	target := filepath.Join(tmpDir, "home", ".vimrc")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/tmp/example", target); err != nil {
		t.Fatal(err)
	}

	h := NewHistoryManager(gdfDir, 512)
	s, err := h.Capture(target)
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	if s == nil || s.Kind != "symlink" {
		t.Fatalf("snapshot = %#v, want symlink snapshot", s)
	}
	if s.LinkTarget != "/tmp/example" {
		t.Fatalf("LinkTarget = %q, want /tmp/example", s.LinkTarget)
	}
}

func TestHistoryManagerEnforcesQuota(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	h := NewHistoryManager(gdfDir, 1) // 1MB

	targetA := filepath.Join(tmpDir, "a")
	targetB := filepath.Join(tmpDir, "b")
	data := make([]byte, 800*1024)
	if err := os.WriteFile(targetA, data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetB, data, 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := h.Capture(targetA); err != nil {
		t.Fatalf("Capture(a) error = %v", err)
	}
	if _, err := h.Capture(targetB); err != nil {
		t.Fatalf("Capture(b) error = %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(gdfDir, ".history"))
	if err != nil {
		t.Fatalf("ReadDir(.history): %v", err)
	}
	if len(entries) > 1 {
		t.Fatalf("expected quota eviction to keep <=1 snapshot, got %d", len(entries))
	}
}
