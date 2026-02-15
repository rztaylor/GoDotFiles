package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeColorMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default empty", input: "", want: "auto"},
		{name: "auto", input: "auto", want: "auto"},
		{name: "always", input: "always", want: "always"},
		{name: "never", input: "never", want: "never"},
		{name: "invalid", input: "rainbow", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeColorMode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeColorMode() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeColorMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintKeyValueLines_NoColor(t *testing.T) {
	originalMode := globalColorMode
	originalKeys := uiHighlightKeyValues
	t.Cleanup(func() {
		globalColorMode = originalMode
		uiHighlightKeyValues = originalKeys
	})

	globalColorMode = "never"
	uiHighlightKeyValues = true
	out := captureStdout(t, func() {
		printKeyValueLines([]keyValue{
			{Key: "Name", Value: "git"},
			{Key: "Aliases", Value: "5"},
		})
	})
	if strings.Contains(out, "\033[") {
		t.Fatalf("unexpected ANSI sequences in no-color mode: %q", out)
	}
	if !strings.Contains(out, "Name") || !strings.Contains(out, "git") {
		t.Fatalf("expected key/value output, got %q", out)
	}
}

func TestPrintKeyValueLines_WithColor(t *testing.T) {
	originalMode := globalColorMode
	originalKeys := uiHighlightKeyValues
	t.Cleanup(func() {
		globalColorMode = originalMode
		uiHighlightKeyValues = originalKeys
	})

	globalColorMode = "always"
	uiHighlightKeyValues = true
	t.Setenv("NO_COLOR", "")
	out := captureStdout(t, func() {
		printKeyValueLines([]keyValue{
			{Key: "Name", Value: "git"},
		})
	})
	if !strings.Contains(out, "\033[36m") {
		t.Fatalf("expected colored key output, got %q", out)
	}
}

func TestConfigureOutputStyle_FromConfig(t *testing.T) {
	originalMode := globalColorMode
	originalHeadings := uiColorSectionHeadings
	originalKeys := uiHighlightKeyValues
	t.Cleanup(func() {
		globalColorMode = originalMode
		uiColorSectionHeadings = originalHeadings
		uiHighlightKeyValues = originalKeys
	})

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	gdfDir := filepath.Join(tmpHome, ".gdf")
	if err := os.MkdirAll(gdfDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(gdfDir, "config.yaml")
	content := `kind: Config/v1
ui:
  color: never
  color_section_headings: false
  highlight_key_values: false
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	globalColorMode = "auto"
	uiColorSectionHeadings = true
	uiHighlightKeyValues = true

	if err := configureOutputStyle(nil); err != nil {
		t.Fatalf("configureOutputStyle() error = %v", err)
	}

	if globalColorMode != "never" {
		t.Fatalf("globalColorMode = %q, want never", globalColorMode)
	}
	if uiColorSectionHeadings {
		t.Fatal("uiColorSectionHeadings = true, want false")
	}
	if uiHighlightKeyValues {
		t.Fatal("uiHighlightKeyValues = true, want false")
	}
}
