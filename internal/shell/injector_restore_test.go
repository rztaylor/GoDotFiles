package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjector_RestoreSourceLine(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)
	rcPath := filepath.Join(home, ".bashrc")

	// Setup existing RC with GDF line
	content := "# Some config\n" +
		"# Added by gdf for shell integration\n" +
		"[ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh\n" +
		"# End config\n"
	if err := os.WriteFile(rcPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	injector := NewInjector()
	err := injector.RestoreSourceLine("~/.aliases", Bash)
	if err != nil {
		t.Fatalf("RestoreSourceLine() error = %v", err)
	}

	// Verify content
	newContent, _ := os.ReadFile(rcPath)
	newStr := string(newContent)

	if strings.Contains(newStr, "[ -f ~/.gdf/generated/init.sh ]") {
		t.Error("Old source line still present")
	}

	expected := "[ -f ~/.aliases ] && source ~/.aliases"
	if !strings.Contains(newStr, expected) {
		t.Errorf("New source line missing. Got:\n%s", newStr)
	}
}
