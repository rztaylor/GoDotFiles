package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage shell aliases",
	Long:  `Add, list, or remove shell aliases within your app bundles.`,
}

var aliasAddCmd = &cobra.Command{
	Use:   "add <name> <command>",
	Short: "Add a shell alias",
	Long: `Add a shell alias to an app bundle.

If --app is specified, the alias is added to that app's bundle.
Otherwise, GDF checks whether the command's first word matches an existing app bundle.
If no match is found, the alias is stored as a global (unassociated) alias.`,
	Args: cobra.ExactArgs(2),
	Example: `  gdf alias add k kubectl
  gdf alias add gco "git checkout" -a git
  gdf alias add ll "ls -la"`,
	RunE: runAliasAdd,
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all aliases",
	Long:  `List all aliases from app bundles and global (unassociated) aliases.`,
	RunE:  runAliasList,
}

var aliasRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a shell alias",
	Long: `Remove a shell alias from its app bundle or global aliases.

Searches all app bundles and the global aliases file to find and remove the alias.`,
	Args: cobra.ExactArgs(1),
	RunE: runAliasRemove,
}

var aliasApp string

func init() {
	rootCmd.AddCommand(aliasCmd)
	aliasCmd.AddCommand(aliasAddCmd)
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasRemoveCmd)

	aliasAddCmd.Flags().StringVarP(&aliasApp, "app", "a", "", "App bundle to add this alias to")
}

func runAliasAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	command := args[1]

	gdfDir := platform.ConfigDir()
	appsDir := filepath.Join(gdfDir, "apps")

	// 1. Determine where to store the alias
	appName := aliasApp
	if appName == "" {
		// Try to auto-detect from existing apps only
		appName = apps.DetectAppFromCommandIfExists(command, appsDir)
	}

	// 2a. No app match → store as global alias
	if appName == "" {
		aliasesPath := filepath.Join(gdfDir, "aliases.yaml")
		ga, err := apps.LoadGlobalAliases(aliasesPath)
		if err != nil {
			return fmt.Errorf("loading global aliases: %w", err)
		}

		if prev, existed := ga.Add(name, command); existed {
			fmt.Printf("Overwriting existing global alias '%s' (was '%s')\n", name, prev)
		}

		if err := ga.Save(aliasesPath); err != nil {
			return fmt.Errorf("saving global aliases: %w", err)
		}

		fmt.Printf("✓ Added global alias '%s=\"%s\"'\n", name, command)
		return nil
	}

	// 2b. App match → store on the app bundle
	if appName != aliasApp && aliasApp == "" {
		fmt.Printf("Detected app: %s\n", appName)
	}

	appPath := filepath.Join(appsDir, appName+".yaml")

	var bundle *apps.Bundle
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		// New app (only when --app was explicitly set)
		bundle = &apps.Bundle{
			Name:        appName,
			Description: fmt.Sprintf("App bundle for %s", appName),
		}
	} else {
		// Existing app
		var err error
		bundle, err = apps.Load(appPath)
		if err != nil {
			return fmt.Errorf("loading app bundle: %w", err)
		}
	}

	// 3. Add to bundle
	if bundle.Shell == nil {
		bundle.Shell = &apps.Shell{}
	}
	if bundle.Shell.Aliases == nil {
		bundle.Shell.Aliases = make(map[string]string)
	}

	if existing, ok := bundle.Shell.Aliases[name]; ok {
		fmt.Printf("Overwriting existing alias '%s' (was '%s')\n", name, existing)
	}

	bundle.Shell.Aliases[name] = command

	// 4. Save
	if err := os.MkdirAll(filepath.Dir(appPath), 0755); err != nil {
		return fmt.Errorf("creating apps directory: %w", err)
	}
	if err := bundle.Save(appPath); err != nil {
		return fmt.Errorf("saving app bundle: %w", err)
	}

	fmt.Printf("✓ Added alias '%s=\"%s\"' to app '%s'\n", name, command, appName)
	return nil
}

func runAliasList(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()
	appsDir := filepath.Join(gdfDir, "apps")

	// 1. Scan all app bundles in apps/ directory
	entries, err := os.ReadDir(appsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading apps directory: %w", err)
	}

	hasAliases := false
	fmt.Println("Aliases by app:")

	// Sort app names for deterministic output
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		appPath := filepath.Join(appsDir, entry.Name())
		bundle, err := apps.Load(appPath)
		if err != nil {
			continue
		}

		if bundle.Shell != nil && len(bundle.Shell.Aliases) > 0 {
			hasAliases = true
			fmt.Printf("\n  %s:\n", bundle.Name)

			// Sort alias names
			names := make([]string, 0, len(bundle.Shell.Aliases))
			for name := range bundle.Shell.Aliases {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				fmt.Printf("    %s = \"%s\"\n", name, bundle.Shell.Aliases[name])
			}
		}
	}

	// 2. Show global (unassociated) aliases
	aliasesPath := filepath.Join(gdfDir, "aliases.yaml")
	ga, err := apps.LoadGlobalAliases(aliasesPath)
	if err != nil {
		return fmt.Errorf("loading global aliases: %w", err)
	}

	if len(ga.Aliases) > 0 {
		hasAliases = true
		fmt.Printf("\n  (unassociated):\n")
		for _, name := range ga.SortedNames() {
			fmt.Printf("    %s = \"%s\"\n", name, ga.Aliases[name])
		}
	}

	if !hasAliases {
		fmt.Println("\n  (none)")
	}

	return nil
}

func runAliasRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	gdfDir := platform.ConfigDir()
	appsDir := filepath.Join(gdfDir, "apps")

	// 1. Search all app bundles in apps/ directory
	entries, err := os.ReadDir(appsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading apps directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		appPath := filepath.Join(appsDir, entry.Name())
		bundle, err := apps.Load(appPath)
		if err != nil {
			continue
		}

		if bundle.Shell != nil && bundle.Shell.Aliases != nil {
			if _, ok := bundle.Shell.Aliases[name]; ok {
				delete(bundle.Shell.Aliases, name)
				if err := bundle.Save(appPath); err != nil {
					return fmt.Errorf("saving app '%s': %w", bundle.Name, err)
				}
				fmt.Printf("✓ Removed alias '%s' from app '%s'\n", name, bundle.Name)
				return nil
			}
		}
	}

	// 2. Search global aliases
	aliasesPath := filepath.Join(gdfDir, "aliases.yaml")
	ga, err := apps.LoadGlobalAliases(aliasesPath)
	if err != nil {
		return fmt.Errorf("loading global aliases: %w", err)
	}

	if ga.Remove(name) {
		if err := ga.Save(aliasesPath); err != nil {
			return fmt.Errorf("saving global aliases: %w", err)
		}
		fmt.Printf("✓ Removed global alias '%s'\n", name)
		return nil
	}

	return fmt.Errorf("alias '%s' not found in any app bundle or global aliases", name)
}
