// Package util provides reusable cross-package helpers for GDF internals.
//
// Key helpers:
//   - WriteFileAtomic: atomically replaces files by writing to a temporary
//     file in the destination directory and renaming into place.
//
// Dependencies:
//   - Standard library filesystem primitives (`os`, `path/filepath`).
package util
