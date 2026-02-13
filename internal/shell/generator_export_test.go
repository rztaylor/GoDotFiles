package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestGenerator_ExportAliases(t *testing.T) {
	bundles := []*apps.Bundle{
		{
			Name: "kubectl",
			Shell: &apps.Shell{
				Aliases: map[string]string{
					"k": "kubectl",
				},
			},
		},
	}
	globalAliases := map[string]string{
		"l": "ls -la",
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, ".aliases")

	g := NewGenerator()
	err := g.ExportAliases(bundles, globalAliases, outputPath)
	if err != nil {
		t.Fatalf("ExportAliases() error = %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# Aliases exported by GDF restore") {
		t.Error("Missing header")
	}
	if !strings.Contains(contentStr, "alias k='kubectl'") {
		t.Error("Missing app alias")
	}
	if !strings.Contains(contentStr, "alias l='ls -la'") {
		t.Error("Missing global alias")
	}
}
