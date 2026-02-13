package apps

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
