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
		"oh-my-zsh":       true,
		"starship":        true,
		"gdf-shell":       true,
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
		{"oh-my-zsh", false},
		{"gdf-shell", false},
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

func TestHighConfidenceCompletionRecipes(t *testing.T) {
	m := New()

	tests := []struct {
		name    string
		bashCmd string
		zshCmd  string
	}{
		{name: "gh", bashCmd: "gh completion -s bash", zshCmd: "gh completion -s zsh"},
		{name: "helm", bashCmd: "helm completion bash", zshCmd: "helm completion zsh"},
		{name: "docker", bashCmd: "docker completion bash", zshCmd: "docker completion zsh"},
		{name: "just", bashCmd: "just --completions bash", zshCmd: "just --completions zsh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := m.Get(tt.name)
			if err != nil {
				t.Fatalf("Get(%q) error = %v", tt.name, err)
			}
			if recipe.Shell == nil || recipe.Shell.Completions == nil {
				t.Fatalf("recipe %q missing shell completions", tt.name)
			}
			if recipe.Shell.Completions.Bash != tt.bashCmd {
				t.Fatalf("recipe %q bash completion = %q, want %q", tt.name, recipe.Shell.Completions.Bash, tt.bashCmd)
			}
			if recipe.Shell.Completions.Zsh != tt.zshCmd {
				t.Fatalf("recipe %q zsh completion = %q, want %q", tt.name, recipe.Shell.Completions.Zsh, tt.zshCmd)
			}
		})
	}
}

func TestRecipeQualityBaselineDotfiles(t *testing.T) {
	m := New()

	tests := map[string][]string{
		"git":      {".gitconfig"},
		"zsh":      {".zshrc"},
		"bat":      {".config/bat/config"},
		"btop":     {".config/btop/btop.conf"},
		"curl":     {".curlrc"},
		"delta":    {".config/delta/config"},
		"direnv":   {".config/direnv/direnvrc"},
		"go":       {".config/go/env"},
		"jq":       {".jq"},
		"nmap":     {".nmaprc"},
		"ripgrep":  {".ripgreprc"},
		"starship": {".config/starship.toml"},
		"wget":     {".wgetrc"},
		"oh-my-zsh": {
			".zshrc",
			".oh-my-zsh/custom/aliases.zsh",
			".oh-my-zsh/custom/functions.zsh",
		},
	}

	for recipeName, targets := range tests {
		t.Run(recipeName, func(t *testing.T) {
			recipe, err := m.Get(recipeName)
			if err != nil {
				t.Fatalf("Get(%q) error = %v", recipeName, err)
			}
			for _, target := range targets {
				if !hasDotfileTarget(recipe, target) {
					t.Fatalf("recipe %q missing dotfile target %q", recipeName, target)
				}
			}
		})
	}
}

func TestOhMyZshRecipeDependencies(t *testing.T) {
	m := New()

	recipe, err := m.Get("oh-my-zsh")
	if err != nil {
		t.Fatalf("Get(%q) error = %v", "oh-my-zsh", err)
	}
	if !containsString(recipe.Dependencies, "zsh") {
		t.Fatalf("recipe %q missing dependency %q", "oh-my-zsh", "zsh")
	}
	if !containsString(recipe.Dependencies, "git") {
		t.Fatalf("recipe %q missing dependency %q", "oh-my-zsh", "git")
	}
}

func TestFDRecipeShellCompatibility(t *testing.T) {
	m := New()

	recipe, err := m.Get("fd")
	if err != nil {
		t.Fatalf("Get(%q) error = %v", "fd", err)
	}
	if recipe.Shell == nil {
		t.Fatalf("recipe %q missing shell config", "fd")
	}
	if recipe.Shell.Aliases["fd"] != "" {
		t.Fatalf("recipe %q should not define an unconditional fd alias", "fd")
	}
	if !hasInitSnippet(recipe, "fd-find-compat-alias") {
		t.Fatalf("recipe %q missing fd compatibility init snippet", "fd")
	}
}

func hasDotfileTarget(recipe *Recipe, target string) bool {
	for _, dotfile := range recipe.Dotfiles {
		if dotfile.Target == target {
			return true
		}
	}
	return false
}

func hasInitSnippet(recipe *Recipe, name string) bool {
	if recipe.Shell == nil {
		return false
	}
	for _, snippet := range recipe.Shell.Init {
		if snippet.Name == name {
			return true
		}
	}
	return false
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
