package library

import (
	"testing"
)

func TestList(t *testing.T) {
	m := New()
	list, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) == 0 {
		t.Error("List() returned empty list")
	}

	expected := map[string]bool{
		"git":             true,
		"zsh":             true,
		"starship":        true,
		"curl":            true,
		"jq":              true,
		"mac-preferences": true,
	}

	for name := range expected {
		found := false
		for _, item := range list {
			if item == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List() missing expected app: %s", name)
		}
	}
}

func TestGet(t *testing.T) {
	m := New()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"git", false},
		{"zsh", false},
		{"nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.Get(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.name {
				t.Errorf("Get() name = %v, want %v", got.Name, tt.name)
			}
			if !tt.wantErr && got.Kind != "Recipe/v1" {
				t.Errorf("Get() kind = %v, want Recipe/v1", got.Kind)
			}
		})
	}
}

func TestAllRecipesValid(t *testing.T) {
	m := New()
	list, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	for _, name := range list {
		t.Run(name, func(t *testing.T) {
			recipe, err := m.Get(name)
			if err != nil {
				t.Errorf("Failed to load recipe %q: %v", name, err)
				return
			}
			if recipe.Name != name {
				t.Errorf("Recipe name mismatch: got %q, want %q", recipe.Name, name)
			}
			if recipe.Kind != "Recipe/v1" {
				t.Errorf("Recipe kind mismatch: got %q, want Recipe/v1", recipe.Kind)
			}
		})
	}
}
