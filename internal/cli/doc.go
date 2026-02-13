// Package cli implements the gdf command-line interface.
//
// # Responsibility
//
// This package is responsible for:
//   - Defining all CLI commands and their flags
//   - Parsing user input and validating arguments
//   - Calling appropriate engine/service functions
//   - Formatting and displaying output to the user
//
// # Key Types
//
//   - rootCmd: The root Cobra command
//   - Various sub-commands: init, add, track, apply, profile, alias, etc.
//
// # Dependencies
//
//   - github.com/spf13/cobra: CLI framework
//   - internal/engine: Core business logic
//   - internal/config: Configuration loading
package cli
