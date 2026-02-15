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

	// Init defines startup snippets to include in generated shell init script.
	// Snippets are emitted in list order for deterministic startup behavior.
	Init []InitSnippet `yaml:"init,omitempty"`
}

// Completions defines commands to generate shell completions.
type Completions struct {
	// Bash is the command to generate bash completions.
	Bash string `yaml:"bash,omitempty"`

	// Zsh is the command to generate zsh completions.
	Zsh string `yaml:"zsh,omitempty"`
}

// InitSnippet defines a shell startup snippet for an app.
type InitSnippet struct {
	// Name uniquely identifies the snippet within an app.
	Name string `yaml:"name"`

	// Common is the default snippet for all shells.
	Common string `yaml:"common,omitempty"`

	// Bash overrides Common for bash shells when set.
	Bash string `yaml:"bash,omitempty"`

	// Zsh overrides Common for zsh shells when set.
	Zsh string `yaml:"zsh,omitempty"`

	// Guard is an optional shell condition checked before executing the snippet.
	Guard string `yaml:"guard,omitempty"`
}
