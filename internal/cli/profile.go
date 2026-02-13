package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/state"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles",
	Long:  `Create, list, and view profiles for organizing your app bundles.`,
}

var profileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new profile",
	Long: `Create a new profile definition.

A profile is a collection of app bundles that can be applied together.
Profiles are stored in profiles/<name>/profile.yaml.`,
	Args: cobra.ExactArgs(1),
	Example: `  gdf profile create work
  gdf profile create home --description "Home environment configuration"`,
	RunE: runProfileCreate,
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Long:  `List all profile definitions in the GDF repository.`,
	RunE:  runProfileList,
}

var profileShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show profile details",
	Long:  `Display detailed information about a specific profile, including apps and includes.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runProfileShow,
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a profile",
	Long: `Delete an existing profile.
	
If the profile contains apps, they will be moved to the 'default' profile.
The 'default' profile cannot be deleted.`,
	Args: cobra.ExactArgs(1),
	RunE: runProfileDelete,
}

var profileRenameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename a profile",
	Long: `Rename an existing profile.

Updates the profile directory, internal name, and references in other profiles.`,
	Args: cobra.ExactArgs(2),
	RunE: runProfileRename,
}

var profileDescription string

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileRenameCmd)

	profileCreateCmd.Flags().StringVar(&profileDescription, "description", "", "Profile description")
}

func runProfileCreate(cmd *cobra.Command, args []string) error {
	profileName := args[0]
	gdfDir := platform.ConfigDir()

	// 1. Check if profile already exists
	profileDir := filepath.Join(gdfDir, "profiles", profileName)
	profilePath := filepath.Join(profileDir, "profile.yaml")

	if _, err := os.Stat(profilePath); err == nil {
		return fmt.Errorf("profile '%s' already exists", profileName)
	}

	// 2. Create profile directory
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}

	// 3. Create profile definition
	profile := &config.Profile{
		Name:        profileName,
		Description: profileDescription,
		Apps:        nil, // Use nil instead of empty slice for cleaner YAML
	}

	if profile.Description == "" {
		profile.Description = fmt.Sprintf("Profile for %s", profileName)
	}

	// 4. Validate and save
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	if err := profile.Save(profilePath); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	fmt.Printf("✓ Created profile '%s' at %s\n", profileName, profileDir)
	return nil
}

func runProfileList(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()
	profilesDir := filepath.Join(gdfDir, "profiles")

	// Load all profiles
	profiles, err := config.LoadAllProfiles(profilesDir)
	if err != nil {
		return fmt.Errorf("loading profiles: %w", err)
	}

	if len(profiles) == 0 {
		fmt.Println("No profiles found.")
		fmt.Println("Create one with: gdf profile create <name>")
		return nil
	}

	fmt.Println("Available profiles:")
	for _, profile := range profiles {
		desc := profile.Description
		if desc == "" {
			desc = "(no description)"
		}
		appCount := len(profile.Apps)
		includeCount := len(profile.Includes)

		fmt.Printf("  • %s - %s\n", profile.Name, desc)
		fmt.Printf("    Apps: %d", appCount)
		if includeCount > 0 {
			fmt.Printf(", Includes: %d", includeCount)
		}
		fmt.Println()
	}

	return nil
}

func runProfileShow(cmd *cobra.Command, args []string) error {
	profileName := "default"
	if len(args) > 0 {
		profileName = args[0]
	}
	gdfDir := platform.ConfigDir()

	// Load profile
	profilePath := filepath.Join(gdfDir, "profiles", profileName, "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("profile '%s' not found", profileName)
		}
		return fmt.Errorf("loading profile: %w", err)
	}

	// Display profile information
	fmt.Printf("Profile: %s\n", profile.Name)
	if profile.Description != "" {
		fmt.Printf("Description: %s\n", profile.Description)
	}
	fmt.Println()

	// Show includes
	if len(profile.Includes) > 0 {
		fmt.Println("Includes:")
		for _, inc := range profile.Includes {
			fmt.Printf("  - %s\n", inc)
		}
		fmt.Println()
	}

	// Show apps
	if len(profile.Apps) > 0 {
		fmt.Printf("Apps (%d):\n", len(profile.Apps))
		for _, app := range profile.Apps {
			fmt.Printf("  - %s\n", app)
		}
	} else {
		fmt.Println("Apps: (none)")
	}

	// Show conditions if any
	if len(profile.Conditions) > 0 {
		fmt.Println()
		fmt.Printf("Conditions (%d):\n", len(profile.Conditions))
		for i, cond := range profile.Conditions {
			fmt.Printf("  %d. If: %s\n", i+1, cond.If)
			if len(cond.IncludeApps) > 0 {
				fmt.Printf("     Include: %v\n", cond.IncludeApps)
			}
			if len(cond.ExcludeApps) > 0 {
				fmt.Printf("     Exclude: %v\n", cond.ExcludeApps)
			}
		}
	}

	return nil
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	profileName := args[0]
	gdfDir := platform.ConfigDir()

	if profileName == "default" {
		return fmt.Errorf("cannot delete 'default' profile")
	}

	profileDir := filepath.Join(gdfDir, "profiles", profileName)
	profilePath := filepath.Join(profileDir, "profile.yaml")

	// Check if profile exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	// Load profile to check for apps
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	// Move apps to default if needed
	if len(profile.Apps) > 0 {
		fmt.Printf("Migrating %d apps to default profile...\n", len(profile.Apps))

		defaultProfilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
		defaultProfile, err := config.LoadProfile(defaultProfilePath)
		if err != nil {
			if os.IsNotExist(err) {
				// Initialize default profile
				if err := os.MkdirAll(filepath.Dir(defaultProfilePath), 0755); err != nil {
					return fmt.Errorf("creating default profile dir: %w", err)
				}
				defaultProfile = &config.Profile{
					Name:        "default",
					Description: "Default profile",
				}
			} else {
				return fmt.Errorf("loading default profile: %w", err)
			}
		}

		// Append apps
		// We should probably avoid duplicates
		existingApps := make(map[string]bool)
		for _, app := range defaultProfile.Apps {
			existingApps[app] = true
		}

		for _, app := range profile.Apps {
			if !existingApps[app] {
				defaultProfile.Apps = append(defaultProfile.Apps, app)
			}
		}

		if err := defaultProfile.Save(defaultProfilePath); err != nil {
			return fmt.Errorf("saving default profile: %w", err)
		}
	}

	// Remove references from other profiles
	profilesDir := filepath.Join(gdfDir, "profiles")
	allProfiles, err := config.LoadAllProfiles(profilesDir)
	if err != nil {
		fmt.Printf("Warning: failed to load profiles for dependency cleanup: %v\n", err)
	} else {
		for _, p := range allProfiles {
			updated := false
			newIncludes := make([]string, 0, len(p.Includes))
			for _, inc := range p.Includes {
				if inc == profileName {
					updated = true
					continue // Skip deleted profile
				}
				newIncludes = append(newIncludes, inc)
			}

			if updated {
				p.Includes = newIncludes
				if err := p.Save(filepath.Join(profilesDir, p.Name, "profile.yaml")); err != nil {
					fmt.Printf("Warning: failed to remove dependency in profile %s: %v\n", p.Name, err)
				} else {
					fmt.Printf("Removed reference in profile '%s'\n", p.Name)
				}
			}
		}
	}

	// Remove from state
	statePath := filepath.Join(gdfDir, "state.yaml")
	st, err := state.Load(statePath)
	if err == nil {
		if st.IsApplied(profileName) {
			st.RemoveProfile(profileName)
			if err := st.Save(statePath); err != nil {
				fmt.Printf("Warning: failed to update state: %v\n", err)
			}
		}
	}

	// Delete profile directory
	if err := os.RemoveAll(profileDir); err != nil {
		return fmt.Errorf("deleting profile directory: %w", err)
	}

	fmt.Printf("✓ Deleted profile '%s'\n", profileName)
	return nil
}

func runProfileRename(cmd *cobra.Command, args []string) error {
	oldName := args[0]
	newName := args[1]
	gdfDir := platform.ConfigDir()

	if oldName == "default" {
		return fmt.Errorf("cannot rename 'default' profile")
	}

	if newName == "default" {
		return fmt.Errorf("cannot rename to 'default' profile")
	}

	// 1. Check if old profile exists
	oldProfileDir := filepath.Join(gdfDir, "profiles", oldName)
	if _, err := os.Stat(oldProfileDir); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' not found", oldName)
	}

	// 2. Check if new profile already exists
	newProfileDir := filepath.Join(gdfDir, "profiles", newName)
	if _, err := os.Stat(newProfileDir); err == nil {
		return fmt.Errorf("profile '%s' already exists", newName)
	}

	// 3. Rename directory
	if err := os.Rename(oldProfileDir, newProfileDir); err != nil {
		return fmt.Errorf("renaming profile directory: %w", err)
	}

	// 4. Update profile.yaml name
	newProfilePath := filepath.Join(newProfileDir, "profile.yaml")
	profile, err := config.LoadProfile(newProfilePath)
	if err != nil {
		return fmt.Errorf("loading moved profile: %w", err)
	}
	profile.Name = newName
	if err := profile.Save(newProfilePath); err != nil {
		return fmt.Errorf("saving updated profile: %w", err)
	}

	// 5. Update dependencies in other profiles
	profilesDir := filepath.Join(gdfDir, "profiles")
	allProfiles, err := config.LoadAllProfiles(profilesDir)
	if err != nil {
		fmt.Printf("Warning: failed to load profiles for dependency update: %v\n", err)
	} else {
		for _, p := range allProfiles {
			updated := false
			for i, inc := range p.Includes {
				if inc == oldName {
					p.Includes[i] = newName
					updated = true
				}
			}
			if updated {
				if err := p.Save(filepath.Join(profilesDir, p.Name, "profile.yaml")); err != nil {
					fmt.Printf("Warning: failed to update dependency in profile %s: %v\n", p.Name, err)
				} else {
					fmt.Printf("Updated reference in profile '%s'\n", p.Name)
				}
			}
		}
	}

	statePath := filepath.Join(gdfDir, "state.yaml")
	st, err := state.Load(statePath)
	if err == nil { // Only update if state exists/loads
		updated := false
		for i := range st.AppliedProfiles {
			if st.AppliedProfiles[i].Name == oldName {
				st.AppliedProfiles[i].Name = newName
				updated = true
			}
		}
		if updated {
			if err := st.Save(statePath); err != nil {
				fmt.Printf("Warning: failed to update state: %v\n", err)
			}
		}
	}

	fmt.Printf("✓ Renamed profile '%s' to '%s'\n", oldName, newName)
	return nil
}
