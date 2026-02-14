package library

import (
	"fmt"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/schema"
)

// Recipe represents a template for an app bundle in the library.
// It uses "Recipe/v1" as its kind.
type Recipe struct {
	schema.TypeMeta `yaml:",inline"`

	// Name is the unique identifier for this recipe.
	Name string `yaml:"name"`

	// Description provides a human-readable explanation of this recipe.
	Description string `yaml:"description,omitempty"`

	// Dependencies lists other app bundles that this recipe depends on.
	Dependencies []string `yaml:"dependencies,omitempty"`

	// Package defines how to install the package.
	Package *apps.Package `yaml:"package,omitempty"`

	// Dotfiles lists configuration files to include (usually empty for recipes, but supported).
	Dotfiles []apps.Dotfile `yaml:"dotfiles,omitempty"`

	// Shell defines shell integration.
	Shell *apps.Shell `yaml:"shell,omitempty"`

	// Hooks defines lifecycle commands to run.
	Hooks *apps.Hooks `yaml:"hooks,omitempty"`

	// Companions lists related apps.
	Companions []string `yaml:"companions,omitempty"`

	// Plugins defines plugin installations.
	Plugins []apps.Plugin `yaml:"plugins,omitempty"`
}

// Validate checks if the recipe is valid.
func (r *Recipe) Validate() error {
	if err := r.ValidateKind("Recipe"); err != nil {
		return err
	}
	if r.Name == "" {
		return fmt.Errorf("recipe name is required")
	}
	return nil
}

// ToBundle converts a recipe into an App Bundle.
// This effectively "instantiates" the recipe.
func (r *Recipe) ToBundle() *apps.Bundle {
	return &apps.Bundle{
		TypeMeta: schema.TypeMeta{
			Kind: "App/v1",
		},
		Name:         r.Name,
		Description:  r.Description,
		Dependencies: r.Dependencies,
		Package:      r.Package,
		Dotfiles:     r.Dotfiles,
		Shell:        r.Shell,
		Hooks:        r.Hooks,
		Companions:   r.Companions,
		Plugins:      r.Plugins,
	}
}
