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
	"fmt"

	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/git"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/state"
	"github.com/rztaylor/GoDotFiles/internal/updater"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gdf",
	Short: "GDF - Go Dotfiles Manager",
	Long: `GDF (Go Dotfiles) is a cross-platform dotfile manager that unifies
packages, configuration files, and shell aliases into coherent "app bundles."

It works on macOS, Linux, and WSL, using Git as the storage backend and
supporting composable profiles for different use cases (work, home, SRE, etc).`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip check for commands that don't require initialization
		if shouldSkipInitCheck(cmd) {
			return nil
		}

		gdfDir := platform.ConfigDir()
		if !git.IsRepository(gdfDir) {
			return fmt.Errorf("GDF repository not initialized at %s. Please run 'gdf init' first.", gdfDir)
		}

		// Auto-update check
		// Skip if command is 'update' to avoid double checks or interfering with manual update
		if cmd.Name() != "update" {
			_ = runAutoUpdateCheck()
		}

		return nil
	},
}

func runAutoUpdateCheck() error {
	cfgPath := platform.ConfigFile()

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return err
	}

	st, err := state.Load(platform.StateFile())
	if err != nil {
		return err
	}

	// Check for update (force=false)
	info, err := updater.CheckForUpdate(cfg, st, false)
	if err != nil {
		return err
	}

	if info != nil {
		return updater.PromptUpdate(info, st)
	}

	return nil
}

func shouldSkipInitCheck(cmd *cobra.Command) bool {
	// Check if the command itself or any of its parents is 'init', 'version', or 'help'
	for c := cmd; c != nil; c = c.Parent() {
		name := c.Name()
		if name == "init" || name == "version" || name == "help" || name == "update" || name == "shell" {
			return true
		}
	}
	return false
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags will be added here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}
