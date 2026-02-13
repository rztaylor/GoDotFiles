// Package cli implements the gdf command-line interface using Cobra.
// It provides all CLI commands for managing dotfiles, apps, profiles, and git operations.
//
// Key commands:
//   - gdf init: Initialize or clone a dotfiles repository
//   - gdf add: Add an app bundle to a profile
//   - gdf track: Track existing dotfiles
//   - gdf apply: Apply profiles to the system
//   - gdf sync: Pull, apply, and push changes
package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gdf",
	Short: "GDF - Go Dotfiles Manager",
	Long: `GDF (Go Dotfiles) is a cross-platform dotfile manager that unifies
packages, configuration files, and shell aliases into coherent "app bundles."

It works on macOS, Linux, and WSL, using Git as the storage backend and
supporting composable profiles for different use cases (work, home, SRE, etc).`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags will be added here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}
