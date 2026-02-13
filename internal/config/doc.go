// Package config handles loading and validating GDF configuration files.
//
// # Responsibility
//
// This package is responsible for:
//   - Loading config.yaml (global settings)
//   - Loading profile.yaml files
//   - Loading app bundle YAML files
//   - Validating configuration schemas
//   - Providing typed access to configuration values
//
// # Key Types
//
//   - Config: Global configuration
//   - Profile: Profile definition
//   - AppBundle: App bundle definition
//
// # Dependencies
//
//   - gopkg.in/yaml.v3: YAML parsing
package config
