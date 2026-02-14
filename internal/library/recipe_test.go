package library

import (
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestRecipe_Validate(t *testing.T) {
	tests := []struct {
		name    string
		recipe  Recipe
		wantErr bool
	}{
		{
			name: "Valid Recipe",
			recipe: Recipe{
				TypeMeta: struct {
					Kind string `yaml:"kind"`
				}{Kind: "Recipe/v1"},
				Name: "git",
			},
			wantErr: false,
		},
		{
			name: "Missing Name",
			recipe: Recipe{
				TypeMeta: struct {
					Kind string `yaml:"kind"`
				}{Kind: "Recipe/v1"},
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid Kind Type",
			recipe: Recipe{
				TypeMeta: struct {
					Kind string `yaml:"kind"`
				}{Kind: "App/v1"},
				Name: "git",
			},
			wantErr: true,
		},
		{
			name: "Invalid Kind Version",
			recipe: Recipe{
				TypeMeta: struct {
					Kind string `yaml:"kind"`
				}{Kind: "Recipe/v2"},
				Name: "git",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.recipe.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Recipe.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecipe_ToBundle(t *testing.T) {
	recipe := Recipe{
		TypeMeta: struct {
			Kind string `yaml:"kind"`
		}{Kind: "Recipe/v1"},
		Name:        "git",
		Description: "Version control",
		Package: &apps.Package{
			Brew: "git",
		},
	}

	bundle := recipe.ToBundle()

	if bundle.Kind != "App/v1" {
		t.Errorf("ToBundle() Kind = %v, want %v", bundle.Kind, "App/v1")
	}
	if bundle.Name != recipe.Name {
		t.Errorf("ToBundle() Name = %v, want %v", bundle.Name, recipe.Name)
	}
	if bundle.Description != recipe.Description {
		t.Errorf("ToBundle() Description = %v, want %v", bundle.Description, recipe.Description)
	}
	if bundle.Package.Brew != recipe.Package.Brew {
		t.Errorf("ToBundle() Package.Brew = %v, want %v", bundle.Package.Brew, recipe.Package.Brew)
	}
}
