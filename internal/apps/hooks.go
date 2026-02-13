package apps

// Hooks defines lifecycle hooks for an app bundle.
type Hooks struct {
	// PreInstall runs before package installation.
	PreInstall []string `yaml:"pre_install,omitempty"`

	// PostInstall runs after package installation.
	PostInstall []string `yaml:"post_install,omitempty"`

	// PreLink runs before dotfile symlinking.
	PreLink []string `yaml:"pre_link,omitempty"`

	// PostLink runs after dotfile symlinking.
	PostLink []string `yaml:"post_link,omitempty"`

	// Apply runs during apply for package-less bundles.
	// Each hook can have an optional condition.
	Apply []ApplyHook `yaml:"apply,omitempty"`
}

// ApplyHook is a hook that runs during apply.
// Used primarily for package-less bundles (e.g., mac-preferences).
type ApplyHook struct {
	// Run is the shell command(s) to execute.
	Run string `yaml:"run"`

	// When is an optional condition expression.
	// Example: "os == 'macos'"
	When string `yaml:"when,omitempty"`
}
