package shell

// ShellType represents the type of shell.
type ShellType int

const (
	// Unknown represents an unknown or unsupported shell.
	Unknown ShellType = iota
	// Bash represents the Bash shell.
	Bash
	// Zsh represents the Zsh shell.
	Zsh
)

// String returns the string representation of the shell type.
func (s ShellType) String() string {
	switch s {
	case Bash:
		return "bash"
	case Zsh:
		return "zsh"
	default:
		return "unknown"
	}
}

// ParseShellType converts a shell name string to a ShellType.
func ParseShellType(shell string) ShellType {
	switch shell {
	case "bash":
		return Bash
	case "zsh":
		return Zsh
	default:
		return Unknown
	}
}
