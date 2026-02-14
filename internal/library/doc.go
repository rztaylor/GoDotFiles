// Package library provides access to the embedded App Library.
//
// It uses Go's embed package to compile a set of curated app recipes
// directly into the binary, allowing users to add common apps without
// needing to download external files.
//
// Key components:
//   - Manager: Handles listing and retrieving embedded recipes
//   - RecipesFS: The embedded filesystem containing YAML files
package library
