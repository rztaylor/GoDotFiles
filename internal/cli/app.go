package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/library"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <app>",
	Short: "Add an app to a profile",
	Long: `Add an app bundle to a profile.

If the app definition (apps/<app>.yaml) does not exist, it will be created.
By default, the app is added to the 'default' profile. Use --to to specify a different profile.`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var removeCmd = &cobra.Command{
	Use:   "remove <app>",
	Short: "Remove an app from a profile",
	Long:  `Remove an app bundle from a profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List apps in a profile",
	Long:  `List all apps in a profile.`,
	RunE:  runList,
}

var (
	targetProfile string
	fromRecipe    bool
)

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(libraryCmd)

	addCmd.Flags().StringVarP(&targetProfile, "profile", "p", "default", "Profile to add app to")
	addCmd.Flags().BoolVar(&fromRecipe, "from-recipe", false, "Use library recipe without prompting")
	removeCmd.Flags().StringVarP(&targetProfile, "profile", "p", "default", "Profile to remove app from")
	listCmd.Flags().StringVarP(&targetProfile, "profile", "p", "default", "Profile to list apps from")
}

func runAdd(cmd *cobra.Command, args []string) error {
	appName := args[0]
	gdfDir := platform.ConfigDir()

	// 1. Check/Create app definition
	appPath := filepath.Join(gdfDir, "apps", appName+".yaml")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		// Check for recipe
		mgr := library.New()
		recipe, err := mgr.Get(appName)
		useRecipe := false

		if err == nil {
			if fromRecipe {
				useRecipe = true
			} else {
				fmt.Printf("Found recipe for '%s' in library.\n", appName)
				fmt.Print("Use this recipe? [Y/n]: ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				if response == "" || response == "y" || response == "yes" {
					useRecipe = true
				}
			}
		}

		if useRecipe {
			fmt.Printf("Using library recipe for '%s'...\n", appName)
			bundle := recipe.ToBundle()
			if err := bundle.Save(appPath); err != nil {
				return fmt.Errorf("saving app bundle: %w", err)
			}
		} else {
			fmt.Printf("App '%s' not found in library.\n", appName)
			fmt.Print("Create new app skeleton? [Y/n]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "" || response == "y" || response == "yes" {
				fmt.Printf("Creating new bundle for '%s'...\n", appName)
				if err := createAppSkeleton(AppName(appName), appPath); err != nil {
					return fmt.Errorf("creating app skeleton: %w", err)
				}
			} else {
				fmt.Println("Aborted.")
				return nil
			}
		}
	} else {
		// App exists - we never override it with a recipe
		fmt.Printf("App '%s' already exists in library (local definition). Using existing configuration.\n", appName)
		// Validate existing app
		if _, err := apps.Load(appPath); err != nil {
			return fmt.Errorf("invalid app definition: %w", err)
		}
	}

	// 2. Load profile
	profileDir := filepath.Join(gdfDir, "profiles", targetProfile)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}

	profilePath := filepath.Join(profileDir, "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Create new profile if it doesn't exist
			profile = &config.Profile{Name: targetProfile}
		} else {
			return fmt.Errorf("loading profile: %w", err)
		}
	}

	// 3. Add app to profile
	if contains(profile.Apps, appName) {
		fmt.Printf("App '%s' is already in profile '%s'\n", appName, targetProfile)
		return nil
	}

	profile.Apps = append(profile.Apps, appName)

	// 4. Save profile
	if err := profile.Save(profilePath); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	fmt.Printf("✓ Added '%s' to profile '%s'\n", appName, targetProfile)
	return nil
}

func runRemove(cmd *cobra.Command, args []string) error {
	appName := args[0]
	gdfDir := platform.ConfigDir()

	// Load profile
	profilePath := filepath.Join(gdfDir, "profiles", targetProfile, "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	// Remove app
	newApps := make([]string, 0, len(profile.Apps))
	found := false
	for _, app := range profile.Apps {
		if app == appName {
			found = true
			continue
		}
		newApps = append(newApps, app)
	}

	if !found {
		fmt.Printf("App '%s' is not in profile '%s'\n", appName, targetProfile)
		return nil
	}

	profile.Apps = newApps

	// Save profile
	if err := profile.Save(profilePath); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	fmt.Printf("✓ Removed '%s' from profile '%s'\n", appName, targetProfile)
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()
	profilePath := filepath.Join(gdfDir, "profiles", targetProfile, "profile.yaml")

	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	fmt.Printf("Apps in profile '%s':\n", targetProfile)
	if len(profile.Apps) == 0 {
		fmt.Println("  (none)")
		return nil
	}

	for _, app := range profile.Apps {
		fmt.Printf("  - %s\n", app)
	}
	return nil
}

func createAppSkeleton(name, path string) error {
	bundle := &apps.Bundle{
		Name:        name,
		Description: fmt.Sprintf("App bundle for %s", name),
	}
	return bundle.Save(path)
}

// AppName sanitizes an app name for use as a filename/identifier.
// Converts to lowercase and replaces spaces and special characters with hyphens.
func AppName(name string) string {
	name = strings.ToLower(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)
	// Remove consecutive hyphens
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")
	return name
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
