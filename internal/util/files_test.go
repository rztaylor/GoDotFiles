package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	if err := WriteFileAtomic(path, []byte("first"), 0644); err != nil {
		t.Fatalf("WriteFileAtomic() first write error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "first" {
		t.Fatalf("content = %q, want %q", string(content), "first")
	}

	if err := WriteFileAtomic(path, []byte("second"), 0600); err != nil {
		t.Fatalf("WriteFileAtomic() second write error = %v", err)
	}

	content, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() after overwrite error = %v", err)
	}
	if string(content) != "second" {
		t.Fatalf("content after overwrite = %q, want %q", string(content), "second")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("mode = %#o, want %#o", info.Mode().Perm(), os.FileMode(0600))
	}
}
