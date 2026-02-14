package library

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/rztaylor/GoDotFiles/internal/schema"
)

//go:embed recipes/*
var RecipesFS embed.FS

// Manager handles the embedded app library.
type Manager struct{}

// NewManager creates a new library manager.
func New() *Manager {
	return &Manager{}
}

// List returns all available recipes in the library.
// It returns a list of app names.
func (m *Manager) List() ([]string, error) {
	var names []string

	err := fs.WalkDir(RecipesFS, "recipes", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		// Parse the file to get the app name (more reliable than filename)
		// But for performance, we might just use filename base.
		// Let's read the file to be safe and accurate.
		data, err := RecipesFS.ReadFile(path)
		if err != nil {
			return err
		}

		var meta struct {
			schema.TypeMeta `yaml:",inline"`
			Name            string `yaml:"name"`
		}
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		if meta.Name != "" {
			names = append(names, meta.Name)
		}
		return nil
	})

	return names, err
}

// Get loads a specific recipe by name.
// It searches for "recipes/*/<name>.yaml"
func (m *Manager) Get(name string) (*Recipe, error) {
	var recipe *Recipe
	found := false

	err := fs.WalkDir(RecipesFS, "recipes", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Match filename with requested name
		// We assume filename matches app name for optimization,
		// but we could also parse every file if strictness is needed.
		// For now, simple filename match: name.yaml
		if strings.TrimSuffix(filepath.Base(path), ".yaml") == name {
			data, err := RecipesFS.ReadFile(path)
			if err != nil {
				return err
			}

			r, err := loadRecipe(data)
			if err != nil {
				return fmt.Errorf("loading recipe %s: %w", path, err)
			}
			recipe = r
			found = true
			return fs.SkipAll // Stop searching
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("recipe not found: %s", name)
	}

	return recipe, nil
}

func loadRecipe(data []byte) (*Recipe, error) {
	var recipe Recipe
	if err := yaml.Unmarshal(data, &recipe); err != nil {
		return nil, err
	}
	if err := recipe.ValidateKind("Recipe"); err != nil {
		return nil, err
	}
	return &recipe, nil
}
