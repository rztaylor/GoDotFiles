package apps

import (
	"github.com/rztaylor/GoDotFiles/internal/schema"
)

// Bundle represents an app bundle - the core unit in GDF.
// A bundle groups together a package, its configuration files (dotfiles),
// A bundle groups together a package, its configuration files (dotfiles),
// shell aliases/functions, and lifecycle hooks.
type Bundle struct {
	schema.TypeMeta `yaml:",inline"`
	// Name is the unique identifier for this app bundle (required).
	Name string `yaml:"name"`

	// Description provides a human-readable explanation of this bundle.
	Description string `yaml:"description,omitempty"`

	// Dependencies lists other app bundles that must be installed first.
	// The engine performs topological sorting during apply.
	Dependencies []string `yaml:"dependencies,omitempty"`

	// Package defines how to install the package. Nil for package-less bundles
	// (e.g., mac-preferences that only run commands).
	Package *Package `yaml:"package,omitempty"`

	// Dotfiles lists configuration files to symlink.
	Dotfiles []Dotfile `yaml:"dotfiles,omitempty"`

	// Shell defines shell integration (aliases, functions, env vars).
	Shell *Shell `yaml:"shell,omitempty"`

	// Hooks defines lifecycle commands to run.
	Hooks *Hooks `yaml:"hooks,omitempty"`

	// Companions lists related apps to suggest to the user.
	Companions []string `yaml:"companions,omitempty"`

	// Plugins defines plugin installations (e.g., krew plugins for kubectl).
	Plugins []Plugin `yaml:"plugins,omitempty"`
}

// Plugin represents a plugin to install for this app.
type Plugin struct {
	// Name is the plugin identifier.
	Name string `yaml:"name"`

	// Install is the command to install this plugin.
	Install string `yaml:"install"`
}
