// Package engine provides the core business logic for GDF.
//
// # Responsibility
//
// This package orchestrates all GDF operations by coordinating other packages:
//   - Profile resolution (handling includes, conditions)
//   - Apply/unapply workflows
//   - State tracking and management
//   - Coordinating packages, dotfiles, and shell generation
//
// # Key Types
//
//   - Engine: Main orchestrator
//   - Linker: Dotfile symlink manager
//   - HistoryManager: Historical snapshot retention manager
//   - Logger: Operation logger used for rollback
//   - State: Current applied state
//   - ApplyResult: Result of an apply operation
//
// # Dependencies
//
//   - internal/apps: App bundle operations
//   - internal/config: Configuration loading
//   - internal/packages: Package installation
//   - internal/shell: Shell script generation
//   - internal/platform: OS detection
package engine
