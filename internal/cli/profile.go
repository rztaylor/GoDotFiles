package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
var profileDeletePurge bool
var profileDeleteMigrateToDefault bool
var profileDeleteLeaveDangling bool
var profileDeleteDryRun bool
var profileDeleteYes bool

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileRenameCmd)

	profileCreateCmd.Flags().StringVar(&profileDescription, "description", "", "Profile description")
	profileDeleteCmd.Flags().BoolVar(&profileDeletePurge, "purge", false, "Purge apps unique to this profile (including managed cleanup)")
	profileDeleteCmd.Flags().BoolVar(&profileDeleteMigrateToDefault, "migrate-to-default", false, "Move this profile's apps into the default profile before deletion")
	profileDeleteCmd.Flags().BoolVar(&profileDeleteLeaveDangling, "leave-dangling", false, "Delete the profile without migrating or purging apps")
	profileDeleteCmd.Flags().BoolVar(&profileDeleteDryRun, "dry-run", false, "Preview profile deletion impact without making changes")
	profileDeleteCmd.Flags().BoolVar(&profileDeleteYes, "yes", false, "Skip confirmation for destructive profile deletion modes")
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
	plat := platform.Detect()

	if profileName == "default" {
		return fmt.Errorf("cannot delete 'default' profile")
	}

	mode, err := resolveProfileDeleteMode()
	if err != nil {
		return err
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

	allProfiles, err := config.LoadAllProfiles(filepath.Join(gdfDir, "profiles"))
	if err != nil {
		return fmt.Errorf("loading profiles: %w", err)
	}

	impact, err := buildProfileDeleteImpact(gdfDir, profileName, profile, allProfiles, plat, mode)
	if err != nil {
		return err
	}
	printProfileDeleteImpact(impact)

	if profileDeleteDryRun {
		fmt.Println("Dry run only. No changes were made.")
		return nil
	}

	if impact.RequiresConfirmation && !profileDeleteYes {
		ok, err := confirmPromptUnsafe("Proceed with profile deletion plan? [y/N]: ")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Aborted.")
			return nil
		}
	}

	switch impact.Mode {
	case profileDeleteModeMigrateToDefault:
		if err := migrateAppsToDefault(gdfDir, profile.Apps); err != nil {
			return err
		}
	case profileDeleteModePurge:
		for _, appName := range impact.PurgeApps {
			plan := impact.PurgePlans[appName]
			if err := executeAppRemovalCleanup(gdfDir, appName, plan); err != nil {
				return err
			}
			if err := purgeAppDefinition(gdfDir, appName); err != nil {
				return err
			}
		}
	case profileDeleteModeLeaveDangling:
		// Intentionally no app movement/cleanup.
	default:
		return fmt.Errorf("unsupported profile delete mode: %s", impact.Mode)
	}

	if err := removeProfileIncludesReferences(gdfDir, profileName, allProfiles); err != nil {
		fmt.Printf("Warning: failed to fully clean include references: %v\n", err)
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

type profileDeleteMode string

const (
	profileDeleteModeMigrateToDefault profileDeleteMode = "migrate-to-default"
	profileDeleteModePurge            profileDeleteMode = "purge"
	profileDeleteModeLeaveDangling    profileDeleteMode = "leave-dangling"
)

type profileDeleteImpact struct {
	Mode                 profileDeleteMode
	RequiresConfirmation bool
	AppsInProfile        []string
	SharedApps           []string
	UniqueApps           []string
	MigrateApps          []string
	DanglingApps         []string
	PurgeApps            []string
	PurgePlans           map[string]*appRemovalPlan
	IncludesAffected     []string
}

func resolveProfileDeleteMode() (profileDeleteMode, error) {
	selected := 0
	mode := profileDeleteModeMigrateToDefault

	if profileDeletePurge {
		selected++
		mode = profileDeleteModePurge
	}
	if profileDeleteMigrateToDefault {
		selected++
		mode = profileDeleteModeMigrateToDefault
	}
	if profileDeleteLeaveDangling {
		selected++
		mode = profileDeleteModeLeaveDangling
	}

	if selected > 1 {
		return "", fmt.Errorf("use only one delete mode: --migrate-to-default, --purge, or --leave-dangling")
	}
	if selected == 0 {
		if globalNonInteractive || !hasInteractiveTerminal() {
			return profileDeleteModeMigrateToDefault, nil
		}
		return chooseProfileDeleteModeInteractive()
	}

	return mode, nil
}

func chooseProfileDeleteModeInteractive() (profileDeleteMode, error) {
	fmt.Println("Choose profile delete strategy:")
	fmt.Println("  1) migrate apps to default")
	fmt.Println("  2) purge unique apps")
	fmt.Println("  3) leave apps dangling")
	input, err := readInteractiveLine("Select [1-3] (default: 1): ")
	if err != nil {
		return "", err
	}
	return parseProfileDeleteModeChoice(input)
}

func parseProfileDeleteModeChoice(input string) (profileDeleteMode, error) {
	choice := strings.TrimSpace(input)
	switch choice {
	case "", "1", "migrate", "migrate-to-default":
		return profileDeleteModeMigrateToDefault, nil
	case "2", "purge":
		return profileDeleteModePurge, nil
	case "3", "leave", "leave-dangling":
		return profileDeleteModeLeaveDangling, nil
	default:
		return "", fmt.Errorf("invalid profile delete strategy %q (use 1, 2, or 3)", choice)
	}
}

func buildProfileDeleteImpact(gdfDir, profileName string, profile *config.Profile, allProfiles []*config.Profile, plat *platform.Platform, mode profileDeleteMode) (*profileDeleteImpact, error) {
	uniqueApps, sharedApps := classifyProfileApps(profileName, profile.Apps, allProfiles)
	impact := &profileDeleteImpact{
		Mode:                 mode,
		RequiresConfirmation: mode == profileDeleteModePurge,
		AppsInProfile:        append([]string(nil), profile.Apps...),
		SharedApps:           sharedApps,
		UniqueApps:           uniqueApps,
		IncludesAffected:     profilesIncludingTarget(profileName, allProfiles),
		PurgePlans:           make(map[string]*appRemovalPlan),
	}

	switch mode {
	case profileDeleteModeMigrateToDefault:
		impact.MigrateApps = append([]string(nil), profile.Apps...)
	case profileDeleteModeLeaveDangling:
		impact.DanglingApps = append([]string(nil), profile.Apps...)
	case profileDeleteModePurge:
		for _, appName := range uniqueApps {
			plan, err := buildAppRemovalPlan(gdfDir, appName, profileName, nil, plat, true)
			if err != nil {
				return nil, err
			}
			impact.PurgeApps = append(impact.PurgeApps, appName)
			impact.PurgePlans[appName] = plan
		}
	}

	return impact, nil
}

func printProfileDeleteImpact(impact *profileDeleteImpact) {
	if impact == nil {
		return
	}

	fmt.Println("Profile delete preview:")
	fmt.Printf("  - mode: %s\n", impact.Mode)
	fmt.Printf("  - apps in profile: %d\n", len(impact.AppsInProfile))
	if len(impact.SharedApps) > 0 {
		fmt.Printf("  - shared apps (kept via other profiles): %s\n", strings.Join(impact.SharedApps, ", "))
	}
	if len(impact.IncludesAffected) > 0 {
		fmt.Printf("  - include references to remove: %s\n", strings.Join(impact.IncludesAffected, ", "))
	}

	switch impact.Mode {
	case profileDeleteModeMigrateToDefault:
		if len(impact.MigrateApps) == 0 {
			fmt.Println("  - migrate apps to default: none")
		} else {
			fmt.Printf("  - migrate apps to default: %s\n", strings.Join(impact.MigrateApps, ", "))
		}
	case profileDeleteModeLeaveDangling:
		if len(impact.DanglingApps) == 0 {
			fmt.Println("  - leave dangling apps: none")
		} else {
			fmt.Printf("  - leave dangling apps: %s\n", strings.Join(impact.DanglingApps, ", "))
		}
	case profileDeleteModePurge:
		if len(impact.PurgeApps) == 0 {
			fmt.Println("  - purge unique apps: none")
			return
		}
		fmt.Printf("  - purge unique apps: %s\n", strings.Join(impact.PurgeApps, ", "))
		for _, appName := range impact.PurgeApps {
			plan := impact.PurgePlans[appName]
			if plan == nil {
				continue
			}
			fmt.Printf("    • %s: unlink=%d", appName, len(plan.UnlinkDotfiles))
			if plan.UninstallPackage {
				fmt.Printf(", uninstall=%s (%s)\n", plan.PackageName, plan.PackageManager)
			} else {
				reason := plan.UninstallSkipReason
				if reason == "" {
					reason = "not applicable"
				}
				fmt.Printf(", uninstall=skipped (%s)\n", reason)
			}
		}
	}
}

func classifyProfileApps(profileName string, profileApps []string, profiles []*config.Profile) (unique []string, shared []string) {
	for _, app := range profileApps {
		sharedElsewhere := false
		for _, p := range profiles {
			if p.Name == profileName {
				continue
			}
			if contains(p.Apps, app) {
				sharedElsewhere = true
				break
			}
		}
		if sharedElsewhere {
			shared = append(shared, app)
		} else {
			unique = append(unique, app)
		}
	}
	return unique, shared
}

func profilesIncludingTarget(profileName string, profiles []*config.Profile) []string {
	out := make([]string, 0)
	for _, p := range profiles {
		for _, inc := range p.Includes {
			if inc == profileName {
				out = append(out, p.Name)
				break
			}
		}
	}
	return out
}

func migrateAppsToDefault(gdfDir string, appsToMigrate []string) error {
	if len(appsToMigrate) == 0 {
		return nil
	}

	fmt.Printf("Migrating %d apps to default profile...\n", len(appsToMigrate))
	defaultProfilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	defaultProfile, err := config.LoadProfile(defaultProfilePath)
	if err != nil {
		if os.IsNotExist(err) {
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

	existing := make(map[string]bool)
	for _, app := range defaultProfile.Apps {
		existing[app] = true
	}
	for _, app := range appsToMigrate {
		if !existing[app] {
			defaultProfile.Apps = append(defaultProfile.Apps, app)
		}
	}
	if err := defaultProfile.Save(defaultProfilePath); err != nil {
		return fmt.Errorf("saving default profile: %w", err)
	}
	return nil
}

func removeProfileIncludesReferences(gdfDir, profileName string, profiles []*config.Profile) error {
	profilesDir := filepath.Join(gdfDir, "profiles")
	var firstErr error
	for _, p := range profiles {
		updated := false
		newIncludes := make([]string, 0, len(p.Includes))
		for _, inc := range p.Includes {
			if inc == profileName {
				updated = true
				continue
			}
			newIncludes = append(newIncludes, inc)
		}

		if !updated {
			continue
		}
		p.Includes = newIncludes
		if err := p.Save(filepath.Join(profilesDir, p.Name, "profile.yaml")); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			fmt.Printf("Warning: failed to remove dependency in profile %s: %v\n", p.Name, err)
			continue
		}
		fmt.Printf("Removed reference in profile '%s'\n", p.Name)
	}
	return firstErr
}

func purgeAppDefinition(gdfDir, appName string) error {
	appPath := filepath.Join(gdfDir, "apps", appName+".yaml")
	if err := os.Remove(appPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing app definition for %s: %w", appName, err)
	}

	dotfilesPath := filepath.Join(gdfDir, "dotfiles", appName)
	if err := os.RemoveAll(dotfilesPath); err != nil {
		return fmt.Errorf("removing dotfiles for %s: %w", appName, err)
	}
	fmt.Printf("✓ Purged app '%s' definition and dotfiles\n", appName)
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
