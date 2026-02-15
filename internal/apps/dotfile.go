package apps

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Dotfile represents a configuration file to be symlinked.
type Dotfile struct {
	// Source is the path relative to ~/.gdf/dotfiles/
	Source string `yaml:"source"`

	// Target is the destination path. Can be a string or TargetMap for
	// platform-specific paths. The ~ prefix is expanded to home directory.
	Target string `yaml:"target,omitempty"`

	// TargetMap provides platform-specific target paths.
	// Used when Target is empty and platform-specific paths are needed.
	TargetMap *TargetMap `yaml:"-"`

	// When is a condition expression for conditional dotfiles.
	// Example: "os == 'macos'" or "hostname =~ 'work-.*'"
	When string `yaml:"when,omitempty"`

	// Template indicates if the file should be rendered as a Go template.
	Template bool `yaml:"template,omitempty"`

	// Secret marks this file as sensitive. Secret files are:
	// - Added to .gitignore
	// - User is warned about committing
	Secret bool `yaml:"secret,omitempty"`
}

// TargetMap provides platform-specific target paths for dotfiles.
type TargetMap struct {
	Default string `yaml:"default,omitempty"`
	Macos   string `yaml:"macos,omitempty"`
	Linux   string `yaml:"linux,omitempty"`
	Wsl     string `yaml:"wsl,omitempty"`
}

// GetTarget returns the target path for the given OS.
// Falls back to Default if the OS-specific path is not set.
func (t *TargetMap) GetTarget(os string) string {
	switch os {
	case "macos":
		if t.Macos != "" {
			return t.Macos
		}
	case "linux":
		if t.Linux != "" {
			return t.Linux
		}
	case "wsl":
		if t.Wsl != "" {
			return t.Wsl
		}
	}
	return t.Default
}

// EffectiveTarget returns the resolved target path for the provided OS.
// If a target map is present, platform-specific values are preferred.
func (d Dotfile) EffectiveTarget(os string) string {
	if d.TargetMap != nil {
		return d.TargetMap.GetTarget(os)
	}
	return d.Target
}

// UnmarshalYAML supports both string and map forms for dotfile.target.
func (d *Dotfile) UnmarshalYAML(node *yaml.Node) error {
	var aux struct {
		Source   string    `yaml:"source"`
		Target   yaml.Node `yaml:"target"`
		When     string    `yaml:"when,omitempty"`
		Template bool      `yaml:"template,omitempty"`
		Secret   bool      `yaml:"secret,omitempty"`
	}

	if err := node.Decode(&aux); err != nil {
		return err
	}

	d.Source = aux.Source
	d.When = aux.When
	d.Template = aux.Template
	d.Secret = aux.Secret
	d.Target = ""
	d.TargetMap = nil

	if aux.Target.Kind == 0 {
		return nil
	}

	switch aux.Target.Kind {
	case yaml.ScalarNode:
		if err := aux.Target.Decode(&d.Target); err != nil {
			return fmt.Errorf("decoding dotfile target as string: %w", err)
		}
	case yaml.MappingNode:
		var targetMap TargetMap
		if err := aux.Target.Decode(&targetMap); err != nil {
			return fmt.Errorf("decoding dotfile target map: %w", err)
		}
		d.TargetMap = &targetMap
	default:
		return fmt.Errorf("invalid dotfile target type: expected string or map")
	}

	return nil
}
