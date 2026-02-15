package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdout pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = original

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	_ = r.Close()
	return buf.String()
}

func TestRunVersion(t *testing.T) {
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	originalVerbose := verbose
	t.Cleanup(func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
		verbose = originalVerbose
	})

	t.Run("prints concise version in default mode", func(t *testing.T) {
		Version = "1.2.3"
		Commit = "abc123"
		Date = "2026-02-15T00:00:00Z"
		verbose = false

		out := captureStdout(t, func() {
			runVersion(nil, nil)
		})

		if !strings.Contains(out, "gdf version 1.2.3") {
			t.Fatalf("output %q missing version line", out)
		}
		if strings.Contains(out, "Latest version:") {
			t.Fatalf("output %q should not perform update checks", out)
		}
	})

	t.Run("prints detailed info in verbose mode", func(t *testing.T) {
		Version = "1.2.3"
		Commit = "abc123"
		Date = "2026-02-15T00:00:00Z"
		verbose = true

		out := captureStdout(t, func() {
			runVersion(nil, nil)
		})

		required := []string{
			"gdf version: 1.2.3",
			"commit: abc123",
			"built at: 2026-02-15T00:00:00Z",
			"go version:",
			"os/arch:",
		}
		for _, want := range required {
			if !strings.Contains(out, want) {
				t.Fatalf("output %q missing %q", out, want)
			}
		}
	})
}
