package apps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGlobalAliases_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.yaml")

	ga, err := LoadGlobalAliases(path)
	if err != nil {
		t.Fatalf("LoadGlobalAliases() error = %v", err)
	}
	if ga.Aliases == nil {
		t.Fatal("Aliases map should be initialized, not nil")
	}
	if len(ga.Aliases) != 0 {
		t.Errorf("expected empty aliases, got %d", len(ga.Aliases))
	}
}

func TestGlobalAliases_SaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.yaml")

	// Save
	ga := &GlobalAliases{
		Aliases: map[string]string{
			"ll": "ls -la",
			"..": "cd ..",
		},
	}
	if err := ga.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("aliases.yaml not created: %v", err)
	}

	// Load
	loaded, err := LoadGlobalAliases(path)
	if err != nil {
		t.Fatalf("LoadGlobalAliases() error = %v", err)
	}

	if loaded.Aliases["ll"] != "ls -la" {
		t.Errorf("alias ll = %q, want 'ls -la'", loaded.Aliases["ll"])
	}
	if loaded.Aliases[".."] != "cd .." {
		t.Errorf("alias .. = %q, want 'cd ..'", loaded.Aliases[".."])
	}
}

func TestGlobalAliases_Add(t *testing.T) {
	ga := &GlobalAliases{Aliases: make(map[string]string)}

	// Add new alias
	prev, existed := ga.Add("ll", "ls -la")
	if existed {
		t.Error("expected existed=false for new alias")
	}
	if prev != "" {
		t.Errorf("expected empty previous, got %q", prev)
	}
	if ga.Aliases["ll"] != "ls -la" {
		t.Errorf("alias ll = %q, want 'ls -la'", ga.Aliases["ll"])
	}

	// Overwrite existing
	prev, existed = ga.Add("ll", "ls -lah")
	if !existed {
		t.Error("expected existed=true for overwrite")
	}
	if prev != "ls -la" {
		t.Errorf("expected previous 'ls -la', got %q", prev)
	}
	if ga.Aliases["ll"] != "ls -lah" {
		t.Errorf("alias ll = %q, want 'ls -lah'", ga.Aliases["ll"])
	}
}

func TestGlobalAliases_Remove(t *testing.T) {
	ga := &GlobalAliases{
		Aliases: map[string]string{
			"ll": "ls -la",
			"..": "cd ..",
		},
	}

	// Remove existing
	if !ga.Remove("ll") {
		t.Error("expected Remove to return true for existing alias")
	}
	if _, ok := ga.Aliases["ll"]; ok {
		t.Error("alias 'll' should be removed")
	}

	// Remove non-existent
	if ga.Remove("nonexistent") {
		t.Error("expected Remove to return false for non-existent alias")
	}

	// Remaining alias intact
	if ga.Aliases[".."] != "cd .." {
		t.Error("alias '..' should still exist")
	}
}

func TestGlobalAliases_SortedNames(t *testing.T) {
	ga := &GlobalAliases{
		Aliases: map[string]string{
			"..":  "cd ..",
			"ll":  "ls -la",
			"cls": "clear",
		},
	}

	names := ga.SortedNames()
	expected := []string{"..", "cls", "ll"}
	if len(names) != len(expected) {
		t.Fatalf("got %d names, want %d", len(names), len(expected))
	}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("names[%d] = %q, want %q", i, name, expected[i])
		}
	}
}
