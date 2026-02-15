package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/engine"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/shell"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore tracked files and system state",
	Long: `Restores all tracked files to their original locations by replacing symlinks with the actual files.
Exports active aliases to a file and updates your shell configuration to source them directly.
This command prepares your system for GDF removal.`,
	RunE: runRestore,
}

var aliasesFile string

func init() {
	recoverCmd.AddCommand(restoreCmd)
	// Default to ~/.aliases (expanded later)
	restoreCmd.Flags().StringVar(&aliasesFile, "aliases-file", "~/.aliases", "Path to export aliases to")
}

func runRestore(cmd *cobra.Command, args []string) error {
	// 1. Warning and Confirmation
	fmt.Println("⚠️  WARNING: This command will restore all tracked files to their original locations,")
	fmt.Println("replacing the managed symlinks. It will also export your current aliases to a file")
	fmt.Println("and update your shell configuration to use that file instead of GDF.")
	fmt.Println("")
	fmt.Print("To proceed, type 'confirmed': ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)

	if input != "confirmed" {
		fmt.Println("Aborted.")
		return nil
	}

	// 2. Load all apps
	gdfDir := platform.ConfigDir()
	appsDir := filepath.Join(gdfDir, "apps")

	bundles, err := apps.LoadAll(appsDir)
	if err != nil {
		return fmt.Errorf("loading apps: %w", err)
	}

	// 3. Restore files
	l := engine.NewLinker("replace")
	fmt.Println("\nRestoring files...")

	count := 0
	for _, bundle := range bundles {
		for _, dotfile := range bundle.Dotfiles {
			if err := l.Restore(dotfile, gdfDir); err != nil {
				fmt.Printf("  x Failed to restore %s: %v\n", dotfile.Target, err)
			} else {
				count++
			}
		}
	}
	fmt.Printf("Processed %d dotfiles.\n", count)

	// 4. Export aliases
	fmt.Println("Exporting aliases...")
	expandedAliasesPath := platform.ExpandPath(aliasesFile)

	// Check if exists
	if _, err := os.Stat(expandedAliasesPath); err == nil {
		fmt.Printf("  File %s already exists. Overwrite? (y/N): ", expandedAliasesPath)
		input, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(input)) != "y" {
			fmt.Println("Skipping alias export.")

			goto UpdateRC
		}
	}

	if err := exportAliases(gdfDir, bundles, expandedAliasesPath); err != nil {
		return err
	}
	fmt.Printf("  ✓ Exported aliases to %s\n", expandedAliasesPath)

UpdateRC:
	// 5. Update RC files
	fmt.Println("Updating shell configuration...")
	injector := shell.NewInjector()

	// Detect shell type
	detectedShell := platform.DetectShell()
	shellType := shell.ParseShellType(detectedShell)

	if shellType != shell.Unknown {
		// Use the flag value (e.g. ~/.aliases) for portability in RC file
		if err := injector.RestoreSourceLine(aliasesFile, shellType); err != nil {
			fmt.Printf("  x Failed to update RC file: %v\n", err)
		} else {
			fmt.Println("  ✓ Updated shell configuration")
		}
	} else {
		fmt.Println("  ! Could not detect supported shell (bash/zsh) to update RC file.")
	}

	fmt.Println("\nRestore complete. You may now safely remove GDF.")
	return nil
}

func exportAliases(gdfDir string, bundles []*apps.Bundle, path string) error {
	// Load global aliases
	ga, err := apps.LoadGlobalAliases(filepath.Join(gdfDir, "aliases.yaml"))
	if err != nil {
		// Verify if it's just missing file or real error
		if !os.IsNotExist(err) {
			fmt.Printf("⚠️  Warning: could not load global aliases: %v\n", err)
		}
		ga = &apps.GlobalAliases{Aliases: make(map[string]string)}
	}

	g := shell.NewGenerator()
	return g.ExportAliases(bundles, ga.Aliases, path)
}
