package platform

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ and environment variables in a path.
func ExpandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home := os.Getenv("HOME")
		if home != "" {
			path = filepath.Join(home, path[1:])
		}
	}

	// Expand environment variables like $HOME, ${VAR}
	path = os.ExpandEnv(path)

	return filepath.Clean(path)
}

// ExpandPathWithHome expands ~ using a specific home directory.
// Useful for testing or when HOME env var is not set.
func ExpandPathWithHome(path, home string) string {
	if path == "" {
		return path
	}

	// Expand ~ to the provided home directory
	if strings.HasPrefix(path, "~") {
		if strings.HasPrefix(path, "~/") || path == "~" {
			path = filepath.Join(home, path[1:])
		}
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return filepath.Clean(path)
}

// NormalizePath converts a path to use forward slashes (for cross-platform consistency).
func NormalizePath(path string) string {
	return filepath.ToSlash(path)
}

// JoinPath joins path elements and cleans the result.
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// ConfigDir returns the GDF config directory (~/.gdf).
func ConfigDir() string {
	return ExpandPath("~/.gdf")
}

// AppsDir returns the apps directory within GDF config.
func AppsDir() string {
	return filepath.Join(ConfigDir(), "apps")
}

// ProfilesDir returns the profiles directory within GDF config.
func ProfilesDir() string {
	return filepath.Join(ConfigDir(), "profiles")
}

// DotfilesDir returns the dotfiles directory within GDF config.
func DotfilesDir() string {
	return filepath.Join(ConfigDir(), "dotfiles")
}

// StateFile returns the path to state.yaml.
func StateFile() string {
	return filepath.Join(ConfigDir(), "state.yaml")
}

// ConfigFile returns the path to config.yaml.
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}
