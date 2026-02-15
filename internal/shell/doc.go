// Package shell generates shell integration scripts for GDF.
//
// # Responsibility
//
// This package handles:
//   - Generating combined alias files from all active apps
//   - Generating function definitions
//   - Setting up environment variables
//   - Emitting app startup/init snippets
//   - Loading shell completions
//   - Optional event-based auto-reload hooks on prompt
//   - Auto-injecting source line into shell RC files
//   - Supporting bash and zsh shells
//
// # Key Types
//
//   - Generator: Main shell script generator that combines shell config from bundles
//   - Injector: Handles auto-injection of source line into .bashrc/.zshrc
//   - ShellType: Enum for supported shells (Bash, Zsh, Unknown)
//
// # Output
//
// Generates ~/.gdf/generated/init.sh which is sourced by the user's shell.
// This file is automatically regenerated during 'gdf apply'.
//
// # Usage
//
// Generate shell integration from bundles:
//
//	generator := shell.NewGenerator()
//	shellType := shell.ParseShellType(platform.DetectShell())
//	err := generator.Generate(bundles, shellType, "~/.gdf/generated/init.sh")
//
// Auto-inject source line during init:
//
//	injector := shell.NewInjector()
//	shellType := shell.ParseShellType(platform.DetectShell())
//	err := injector.InjectSourceLine(shellType)
//
// # Security
//
// The injector creates a backup (.gdf.backup) before modifying RC files
// and prevents duplicate injection by checking if the source line already exists.
package shell
