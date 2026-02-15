// Package cli implements the gdf command-line interface using Cobra.
// It provides all CLI commands for managing dotfiles, apps, profiles, and git operations.
//
// Key commands:
//   - gdf init: Initialize or clone a dotfiles repository
//   - gdf app add: Add an app bundle to a profile
//   - gdf app track: Track existing dotfiles
//   - gdf apply: Apply profiles to the system
//   - gdf sync: Pull, commit, and push repository changes
package cli

import (
	"fmt"
	"os"

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
		if err := configureOutputStyle(cmd); err != nil {
			return err
		}

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
	if globalNonInteractive || !hasInteractiveTerminal() {
		return nil
	}

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

func hasInteractiveTerminal() bool {
	in, err := os.Stdin.Stat()
	if err != nil || in.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	out, err := os.Stdout.Stat()
	if err != nil || out.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	return true
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
	rootCmd.PersistentFlags().BoolVarP(&globalVerbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&globalColorMode, "color", "auto", "Color mode: auto, always, never")
	rootCmd.PersistentFlags().BoolVar(&globalYes, "yes", false, "Automatically approve safe interactive prompts")
	rootCmd.PersistentFlags().BoolVar(&globalNonInteractive, "non-interactive", false, "Disable interactive prompts and fail when confirmation is required")
}
