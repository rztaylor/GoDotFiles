package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/engine"
	"github.com/rztaylor/GoDotFiles/internal/library"
	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <app>",
	Short: "Add an app to a profile",
	Long: `Add an app bundle to a profile.

If the app definition (apps/<app>.yaml) does not exist, it will be created.
By default, the app is added to the 'default' profile. Use --profile to specify a different profile.`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage app bundles and recipes",
	Long:  `Manage app bundles and recipe-driven app workflows.`,
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

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Archive or delete orphaned app definitions",
	Long:  `Find app definitions that are not referenced by any profile and archive or delete them.`,
	RunE:  runPrune,
}

var (
	targetProfile   string
	fromRecipe      bool
	addInteractive  bool
	removeUninstall bool
	removeYes       bool
	removeDryRun    bool
	pruneDryRun     bool
	pruneDelete     bool
	pruneJSON       bool
	pruneYes        bool
)

func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(addCmd)
	appCmd.AddCommand(removeCmd)
	appCmd.AddCommand(listCmd)
	appCmd.AddCommand(pruneCmd)
	appCmd.AddCommand(libraryCmd)

	addCmd.Flags().StringVarP(&targetProfile, "profile", "p", "default", "Profile to add app to")
	addCmd.Flags().BoolVar(&fromRecipe, "from-recipe", false, "Use library recipe without prompting")
	addCmd.Flags().BoolVar(&addInteractive, "interactive", false, "Enable interactive recipe suggestions and dependency prompts")
	removeCmd.Flags().StringVarP(&targetProfile, "profile", "p", "default", "Profile to remove app from")
	removeCmd.Flags().BoolVar(&removeUninstall, "uninstall", false, "Also unlink managed dotfiles and uninstall the app package when no profiles reference the app")
	removeCmd.Flags().BoolVar(&removeDryRun, "dry-run", false, "Preview removal actions without making changes")
	removeCmd.Flags().BoolVar(&removeYes, "yes", false, "Skip confirmation prompt for uninstall/unlink cleanup")
	listCmd.Flags().StringVarP(&targetProfile, "profile", "p", "default", "Profile to list apps from")
	pruneCmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "Preview prune actions without making changes")
	pruneCmd.Flags().BoolVar(&pruneDelete, "delete", false, "Permanently delete orphaned app definitions instead of archiving them")
	pruneCmd.Flags().BoolVar(&pruneJSON, "json", false, "Output prune results as JSON")
	pruneCmd.Flags().BoolVar(&pruneYes, "yes", false, "Skip confirmation for permanent delete mode")
}

func runAdd(cmd *cobra.Command, args []string) error {
	appName := args[0]
	gdfDir := platform.ConfigDir()
	if addInteractive {
		if err := maybeSuggestRecipes(appName); err != nil {
			return err
		}
	}

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
				confirmed, err := confirmPromptDefaultYes("Use this recipe? [Y/n]: ")
				if err != nil {
					return err
				}
				useRecipe = confirmed
			}
		}

		if useRecipe {
			fmt.Printf("Using library recipe for '%s'...\n", appName)
			bundle := recipe.ToBundle()
			if err := bundle.Save(appPath); err != nil {
				return fmt.Errorf("saving app bundle: %w", err)
			}
			if addInteractive {
				if err := maybeIncludeRecipeDependencies(bundle, targetProfile, gdfDir); err != nil {
					return err
				}
			}
		} else {
			fmt.Printf("App '%s' not found in library.\n", appName)
			confirmed, err := confirmPromptDefaultYes("Create new app skeleton? [Y/n]: ")
			if err != nil {
				return err
			}
			if confirmed {
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

func maybeSuggestRecipes(appName string) error {
	mgr := library.New()
	names, err := mgr.List()
	if err != nil {
		return nil
	}

	suggestions := make([]string, 0, 3)
	lower := strings.ToLower(appName)
	for _, name := range names {
		n := strings.ToLower(name)
		if strings.Contains(n, lower) || strings.Contains(lower, n) {
			suggestions = append(suggestions, name)
		}
		if len(suggestions) == 3 {
			break
		}
	}
	if len(suggestions) > 0 {
		fmt.Printf("Recipe suggestions: %s\n", strings.Join(suggestions, ", "))
	}
	return nil
}

func maybeIncludeRecipeDependencies(bundle *apps.Bundle, profileName, gdfDir string) error {
	if bundle == nil || len(bundle.Dependencies) == 0 {
		return nil
	}

	fmt.Printf("Recipe dependencies for '%s': %s\n", bundle.Name, strings.Join(bundle.Dependencies, ", "))
	includeDeps, err := confirmPromptDefaultYes("Add these dependencies to the profile now? [Y/n]: ")
	if err != nil {
		return err
	}
	if !includeDeps {
		return nil
	}

	for _, dep := range bundle.Dependencies {
		if err := addAppToProfile(gdfDir, profileName, dep); err != nil {
			return err
		}
	}
	return nil
}

func runRemove(cmd *cobra.Command, args []string) error {
	appName := args[0]
	gdfDir := platform.ConfigDir()
	plat := platform.Detect()

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

	removePlan, err := buildAppRemovalPlan(gdfDir, appName, targetProfile, newApps, plat, removeUninstall)
	if err != nil {
		return err
	}
	printAppRemovalPlan(removePlan)

	if removeDryRun {
		fmt.Println("Dry run only. No changes were made.")
		return nil
	}

	if removePlan.RequiresConfirmation && !removeYes {
		ok, err := confirmPromptUnsafe("Proceed with app removal plan? [y/N]: ")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Save profile
	if err := profile.Save(profilePath); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	fmt.Printf("✓ Removed '%s' from profile '%s'\n", appName, targetProfile)

	if err := executeAppRemovalCleanup(gdfDir, appName, removePlan); err != nil {
		return err
	}
	printDanglingAppCleanupGuidance(gdfDir, appName)

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

type appRemovalPlan struct {
	RemoveFromProfile    bool
	RequiresConfirmation bool
	UninstallRequested   bool
	AppStillReferenced   bool
	ReferencedByProfiles []string
	UnlinkDotfiles       []apps.Dotfile
	PackageManager       string
	PackageName          string
	UninstallPackage     bool
	UninstallSkipReason  string
	UnlinkSkipReason     string
}

func buildAppRemovalPlan(gdfDir, appName, profileName string, profileAppsAfterRemoval []string, plat *platform.Platform, uninstallRequested bool) (*appRemovalPlan, error) {
	plan := &appRemovalPlan{
		RemoveFromProfile:    true,
		UninstallRequested:   uninstallRequested,
		RequiresConfirmation: uninstallRequested,
	}

	if !uninstallRequested {
		return plan, nil
	}

	profiles, err := config.LoadAllProfiles(filepath.Join(gdfDir, "profiles"))
	if err != nil {
		return nil, fmt.Errorf("loading profiles: %w", err)
	}

	referencedBy := referencedProfilesForApp(profiles, appName, profileName, profileAppsAfterRemoval)
	if len(referencedBy) > 0 {
		plan.AppStillReferenced = true
		plan.ReferencedByProfiles = referencedBy
		plan.UnlinkSkipReason = "app remains referenced by other profiles"
		plan.UninstallSkipReason = "app remains referenced by other profiles"
		return plan, nil
	}

	appPath := filepath.Join(gdfDir, "apps", appName+".yaml")
	bundle, err := apps.Load(appPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			plan.UnlinkSkipReason = "app definition not found"
			plan.UninstallSkipReason = "app definition not found"
			return plan, nil
		}
		return nil, fmt.Errorf("loading app bundle: %w", err)
	}

	unlinkDotfiles, err := collectManagedDotfilesForCurrentPlatform(bundle, plat)
	if err != nil {
		return nil, err
	}
	if len(unlinkDotfiles) == 0 {
		plan.UnlinkSkipReason = "no managed dotfiles for current platform"
	}
	plan.UnlinkDotfiles = unlinkDotfiles

	if bundle.Package == nil {
		plan.UninstallSkipReason = "app has no package definition"
		return plan, nil
	}

	mgr := packages.ForPlatform(plat)
	plan.PackageManager = mgr.Name()
	if mgr.Name() == "none" {
		plan.UninstallSkipReason = "no package manager detected"
		return plan, nil
	}

	pkgName, defined := bundle.Package.ResolveName(mgr.Name())
	if !defined || pkgName == "" {
		plan.UninstallSkipReason = fmt.Sprintf("app has no package mapping for manager '%s'", mgr.Name())
		return plan, nil
	}
	plan.PackageName = pkgName

	referencedApps := referencedAppsAfterRemoval(profiles, profileName, profileAppsAfterRemoval)
	unique, err := packageUniqueToApp(gdfDir, appName, mgr.Name(), pkgName, referencedApps)
	if err != nil {
		return nil, err
	}
	if !unique {
		plan.UninstallSkipReason = fmt.Sprintf("package '%s' is shared by another referenced app", pkgName)
		return plan, nil
	}

	plan.UninstallPackage = true
	return plan, nil
}

func executeAppRemovalCleanup(gdfDir, appName string, plan *appRemovalPlan) error {
	if plan == nil || !plan.UninstallRequested || plan.AppStillReferenced {
		return nil
	}

	// Cleanup managed symlinks with rollback snapshots.
	linkedRemoved := 0
	logger := engine.NewLogger(false)
	if len(plan.UnlinkDotfiles) > 0 {
		cfg, err := config.LoadConfig(filepath.Join(gdfDir, "config.yaml"))
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		linker := engine.NewLinker(cfg.ConflictResolution.DotfilesDefault())
		linker.SetHistoryManager(engine.NewHistoryManager(gdfDir, cfg.History.MaxSizeMBDefault()))

		for _, dotfile := range plan.UnlinkDotfiles {
			snapshot, err := linker.UnlinkManaged(dotfile, gdfDir)
			if err != nil {
				return fmt.Errorf("unlinking %s: %w", dotfile.Target, err)
			}

			target := platform.ExpandPath(dotfile.Target)
			details := map[string]string{
				"source":     dotfile.Source,
				"app":        appName,
				"source_abs": filepath.Join(gdfDir, "dotfiles", dotfile.Source),
			}
			if snapshot != nil {
				details["snapshot_id"] = snapshot.ID
				details["snapshot_path"] = snapshot.Path
				details["snapshot_kind"] = snapshot.Kind
				details["snapshot_link_target"] = snapshot.LinkTarget
				details["snapshot_mode"] = fmt.Sprintf("%#o", uint32(snapshot.Mode.Perm()))
				details["snapshot_checksum"] = snapshot.Checksum
				details["snapshot_size_bytes"] = fmt.Sprintf("%d", snapshot.SizeBytes)
				details["snapshot_captured_at"] = snapshot.CapturedAt.Format("2006-01-02T15:04:05.999999999Z07:00")
			}
			logger.Log("link", target, details)
			linkedRemoved++
		}
	}

	if plan.UninstallPackage {
		mgr := packages.ForPlatform(platform.Detect())
		if err := mgr.Uninstall(plan.PackageName); err != nil {
			return fmt.Errorf("uninstalling package %s via %s: %w", plan.PackageName, mgr.Name(), err)
		}
		fmt.Printf("✓ Uninstalled package '%s' via %s\n", plan.PackageName, mgr.Name())
	}

	if linkedRemoved > 0 {
		logPath, err := logger.Save(gdfDir)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to save unlink operation log: %v\n", err)
		} else if logPath != "" {
			fmt.Printf("📝 Logged unlink operations for rollback: %s\n", logPath)
		}
		fmt.Printf("✓ Unlinked %d managed dotfile(s)\n", linkedRemoved)
	}

	return nil
}

func printAppRemovalPlan(plan *appRemovalPlan) {
	if plan == nil {
		return
	}
	fmt.Println("Removal plan:")
	fmt.Println("  - remove app from selected profile: yes")

	if !plan.UninstallRequested {
		fmt.Println("  - uninstall/unlink cleanup: not requested")
		return
	}

	if plan.AppStillReferenced {
		fmt.Printf("  - app is still referenced by profiles: %s\n", strings.Join(plan.ReferencedByProfiles, ", "))
		fmt.Println("  - unlink managed dotfiles: skipped")
		fmt.Println("  - package uninstall: skipped")
		return
	}

	if len(plan.UnlinkDotfiles) == 0 {
		fmt.Printf("  - unlink managed dotfiles: skipped (%s)\n", plan.UnlinkSkipReason)
	} else {
		fmt.Printf("  - unlink managed dotfiles: %d target(s)\n", len(plan.UnlinkDotfiles))
	}

	if plan.UninstallPackage {
		fmt.Printf("  - uninstall package: %s via %s\n", plan.PackageName, plan.PackageManager)
	} else {
		reason := plan.UninstallSkipReason
		if reason == "" {
			reason = "not applicable"
		}
		fmt.Printf("  - uninstall package: skipped (%s)\n", reason)
	}
}

func collectManagedDotfilesForCurrentPlatform(bundle *apps.Bundle, plat *platform.Platform) ([]apps.Dotfile, error) {
	if bundle == nil {
		return nil, nil
	}
	out := make([]apps.Dotfile, 0, len(bundle.Dotfiles))
	for _, dotfile := range bundle.Dotfiles {
		if dotfile.When != "" {
			match, err := config.EvaluateCondition(dotfile.When, plat)
			if err != nil {
				return nil, fmt.Errorf("evaluating condition for dotfile %s in app %s: %w", dotfile.Source, bundle.Name, err)
			}
			if !match {
				continue
			}
		}
		effectiveTarget := dotfile.EffectiveTarget(plat.OS)
		if effectiveTarget == "" {
			continue
		}
		d := dotfile
		d.Target = effectiveTarget
		out = append(out, d)
	}
	return out, nil
}

func referencedProfilesForApp(profiles []*config.Profile, appName, profileBeingEdited string, profileAppsAfterRemoval []string) []string {
	names := make([]string, 0)
	for _, p := range profiles {
		appsInProfile := p.Apps
		if p.Name == profileBeingEdited {
			appsInProfile = profileAppsAfterRemoval
		}
		if contains(appsInProfile, appName) {
			names = append(names, p.Name)
		}
	}
	return names
}

func referencedAppsAfterRemoval(profiles []*config.Profile, profileBeingEdited string, profileAppsAfterRemoval []string) map[string]bool {
	out := make(map[string]bool)
	for _, p := range profiles {
		appsInProfile := p.Apps
		if p.Name == profileBeingEdited {
			appsInProfile = profileAppsAfterRemoval
		}
		for _, app := range appsInProfile {
			out[app] = true
		}
	}
	return out
}

func packageUniqueToApp(gdfDir, appName, managerName, packageName string, referencedApps map[string]bool) (bool, error) {
	bundles, err := apps.LoadAll(filepath.Join(gdfDir, "apps"))
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, fmt.Errorf("loading app bundles: %w", err)
	}

	for _, bundle := range bundles {
		if bundle.Name == appName || !referencedApps[bundle.Name] || bundle.Package == nil {
			continue
		}
		if otherPkg, ok := bundle.Package.ResolveName(managerName); ok && otherPkg == packageName {
			return false, nil
		}
	}
	return true, nil
}
