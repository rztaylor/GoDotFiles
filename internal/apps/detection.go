package apps

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectAppFromPath tries to determine the app name from a file path.
func DetectAppFromPath(path string) string {
	base := filepath.Base(path)

	// Known mappings
	mappings := map[string]string{
		".gitconfig":    "git",
		".gitignore":    "git",
		".zshrc":        "zsh",
		".bashrc":       "bash",
		".bash_profile": "bash",
		".vimrc":        "vim",
		"init.vim":      "nvim",
		"config.fish":   "fish",
		"starship.toml": "starship",
		".tmux.conf":    "tmux",
	}

	if app, ok := mappings[base]; ok {
		return app
	}

	// Check parent dir for .config/APP/file patterns
	dir := filepath.Base(filepath.Dir(path))
	parent := filepath.Base(filepath.Dir(filepath.Dir(path)))
	if parent == ".config" {
		return dir
	}

	// Default: use filename
	name := base
	if strings.HasPrefix(name, ".") {
		// Dotfile: .zshrc -> zshrc
		name = strings.TrimPrefix(name, ".")
	} else {
		// Normal file: config.toml -> config
		ext := filepath.Ext(name)
		if ext != "" {
			name = strings.TrimSuffix(name, ext)
		}
	}

	// Sanitize
	name = strings.ToLower(name)
	return name
}

// DetectAppFromCommand tries to determine the app name from a command string.
func DetectAppFromCommand(cmd string) string {
	// Simple heuristic: first word of command
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "unknown"
	}

	prog := parts[0]
	// Handle absolute paths: /usr/bin/git -> git
	base := filepath.Base(prog)

	// Sanitize
	return strings.ToLower(base)
}

// DetectAppFromCommandIfExists tries to determine the app name from a command
// string, but only returns a match if a corresponding app bundle already exists
// in the apps directory. Returns an empty string if no existing app matches.
func DetectAppFromCommandIfExists(cmd string, appsDir string) string {
	candidate := DetectAppFromCommand(cmd)
	if candidate == "" || candidate == "unknown" {
		return ""
	}

	appPath := filepath.Join(appsDir, candidate+".yaml")
	if _, err := os.Stat(appPath); err != nil {
		return ""
	}

	return candidate
}
