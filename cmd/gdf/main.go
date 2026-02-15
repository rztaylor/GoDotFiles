// Package main is the entry point for the gdf CLI application.
// GDF (Go Dotfiles) is a cross-platform dotfile manager that unifies
// packages, configuration files, and shell aliases into coherent "app bundles."
package main

import (
	"fmt"
	"os"

	"github.com/rztaylor/GoDotFiles/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCode(err))
	}
}
