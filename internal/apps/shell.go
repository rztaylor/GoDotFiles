package apps

// Shell defines shell integration for an app bundle.
type Shell struct {
	// Aliases maps alias names to commands.
	// Example: {"k": "kubectl", "kgp": "kubectl get pods"}
	Aliases map[string]string `yaml:"aliases,omitempty"`

	// Functions maps function names to their bodies.
	Functions map[string]string `yaml:"functions,omitempty"`

	// Env maps environment variable names to values.
	Env map[string]string `yaml:"env,omitempty"`

	// Completions defines shell completion generation commands.
	Completions *Completions `yaml:"completions,omitempty"`
}

// Completions defines commands to generate shell completions.
type Completions struct {
	// Bash is the command to generate bash completions.
	Bash string `yaml:"bash,omitempty"`

	// Zsh is the command to generate zsh completions.
	Zsh string `yaml:"zsh,omitempty"`
}
