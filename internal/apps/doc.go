// Package apps manages app bundles - the core unit in GDF.
//
// # Responsibility
//
// This package handles:
//   - Loading and saving app bundle definitions
//   - Managing app relationships (companions, plugins)
//   - Auto-detection of apps from paths and commands
//   - App bundle validation
//
// # Key Types
//
//   - Bundle: An app bundle (package + dotfiles + aliases)
//   - Companion: A companion app relationship
//   - Plugin: A plugin definition
//
// # Dependencies
//
//   - internal/config: For loading YAML definitions
package apps
