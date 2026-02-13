// Package util provides shared utility functions for GDF.
//
// # Responsibility
//
// This package provides common utilities used across packages:
//   - File operations (copy, move, symlink)
//   - String helpers
//   - Path utilities
//   - Error wrapping helpers
//
// # Key Functions
//
//   - CopyFile: Copy a file with permissions
//   - Symlink: Create symlinks with backup
//   - ExpandPath: Expand ~ and environment variables
//   - FileExists: Check if file exists
package util
