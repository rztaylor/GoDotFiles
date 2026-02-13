package cli

import (
	"fmt"
	"path/filepath"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/engine"
	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/shell"
	"github.com/rztaylor/GoDotFiles/internal/state"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply <profiles...>",
	Short: "Apply one or more profiles",
	Long: `Apply profiles to the system.

This command will:
  1. Resolve profile dependencies (includes)
  2. Resolve app dependencies
  3. Install packages (if package manager available)
  4. Link dotfiles with conflict resolution
  5. Run apply hooks for package-less bundles
  6. Generate shell integration scripts

All operations are logged to ~/.gdf/.operations/ for potential rollback.`,
	Example: `  gdf apply base work
  gdf apply --dry-run sre
  gdf apply base`,
	Args: cobra.MinimumNArgs(1),
	RunE: runApply,
}

var applyDryRun bool

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "Show what would be done without making changes")
}

func runApply(cmd *cobra.Command, args []string) error {
	profileNames := args
	gdfDir := platform.ConfigDir()
	plat := platform.Detect()

	// Load config for conflict strategy
	cfg, err := config.LoadConfig(filepath.Join(gdfDir, "config.yaml"))
	if err != nil {
		// Fallback to default if config missing
		cfg = &config.Config{
			ConflictResolution: &config.ConflictResolution{
				Dotfiles: "backup_and_replace",
			},
		}
	}

	// Initialize operation logger
	logger := engine.NewLogger(applyDryRun)

	if applyDryRun {
		fmt.Println("🔍 Dry run mode - no changes will be made")
		fmt.Println()
	}

	// Phase 1: Resolve profile dependencies
	fmt.Printf("Resolving profile dependencies for: %v\n", profileNames)

	profilesDir := filepath.Join(gdfDir, "profiles")
	allProfiles, err := config.LoadAllProfiles(profilesDir)
	if err != nil {
		return fmt.Errorf("loading profiles: %w", err)
	}

	profileMap := config.ProfileMap(allProfiles)
	resolvedProfiles, err := config.ResolveProfiles(profileNames, profileMap, plat)
	if err != nil {
		return fmt.Errorf("resolving profiles: %w", err)
	}

	fmt.Printf("✓ Profiles to apply (in order): ")
	for i, p := range resolvedProfiles {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(p.Name)
	}
	fmt.Println()

	// Phase 2: Collect all apps from profiles
	appNames := make(map[string]bool)
	for _, profile := range resolvedProfiles {
		for _, app := range profile.Apps {
			appNames[app] = true
		}
	}

	if len(appNames) == 0 {
		fmt.Println("No apps to apply")
		return nil
	}

	// Phase 3: Load all app bundles
	appsDir := filepath.Join(gdfDir, "apps")
	allBundles := make(map[string]*apps.Bundle)
	for appName := range appNames {
		appPath := filepath.Join(appsDir, appName+".yaml")
		bundle, err := apps.Load(appPath)
		if err != nil {
			fmt.Printf("⚠️  Warning: skipping app '%s': %v\n", appName, err)
			continue
		}
		allBundles[appName] = bundle
	}

	// Phase 4: Resolve app dependencies
	fmt.Println("\nResolving app dependencies...")
	appNamesSlice := make([]string, 0, len(allBundles))
	for name := range allBundles {
		appNamesSlice = append(appNamesSlice, name)
	}

	resolvedApps, err := apps.ResolveApps(appNamesSlice, allBundles)
	if err != nil {
		return fmt.Errorf("resolving app dependencies: %w", err)
	}

	fmt.Printf("✓ Apps to process (in order): ")
	for i, app := range resolvedApps {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(app.Name)
	}
	fmt.Println() // Removed the extra newline from fmt.Println("\n")

	// Phase 5: Apply each app
	pkgMgr := packages.ForPlatform(plat)
	linker := engine.NewLinker(cfg.ConflictResolution.Dotfiles)

	for _, bundle := range resolvedApps {
		fmt.Printf("📦 Processing app: %s\n", bundle.Name)

		// 5a. Install package (if defined)
		if bundle.Package != nil {
			// Get package name using platform-specific resolution
			pkgName, defined := bundle.Package.ResolveName(pkgMgr.Name())

			// If not defined for this specific manager, check if we should fallback or skip
			if !defined {
				fmt.Printf("      ⚠️  App '%s' is not configured for package manager '%s'. Skipping package install.\n", bundle.Name, pkgMgr.Name())
				goto SkipPackageInstall
			}

			fmt.Printf("   Package: %s (via %s)\n", pkgName, pkgMgr.Name())

			if pkgMgr.Name() != "none" {
				if !applyDryRun {
					if err := pkgMgr.Install(pkgName); err != nil {
						return fmt.Errorf("installing package %s: %w", pkgName, err)
					}
				}
				logger.Log("package_install", pkgName, map[string]string{
					"manager": pkgMgr.Name(),
					"app":     bundle.Name,
				})
			} else {
				fmt.Printf("      ⏭️  Skipping (no package manager)\n")
			}
		}

	SkipPackageInstall:
		// 5b. Link dotfiles
		if len(bundle.Dotfiles) > 0 {
			fmt.Printf("   Dotfiles: %d file(s)\n", len(bundle.Dotfiles))
			for _, dotfile := range bundle.Dotfiles {
				if !applyDryRun {
					if err := linker.Link(dotfile, gdfDir); err != nil {
						return fmt.Errorf("linking %s: %w", dotfile.Source, err)
					}
				}
				fmt.Printf("      ✓ %s → %s\n", dotfile.Target, dotfile.Source)
				logger.Log("link", dotfile.Target, map[string]string{
					"source": dotfile.Source,
					"app":    bundle.Name,
				})
			}
		}

		// 5c. Run apply hooks (for package-less bundles)
		if bundle.Hooks != nil && len(bundle.Hooks.Apply) > 0 {
			fmt.Printf("   Apply hooks: %d command(s)\n", len(bundle.Hooks.Apply))
			for _, hook := range bundle.Hooks.Apply {
				fmt.Printf("      • %s\n", hook.Run)
				if !applyDryRun {
					// TODO: Actually run the hook command
					// For now, just log it
				}
				logger.Log("hook_run", hook.Run, map[string]string{
					"type": "apply",
					"app":  bundle.Name,
					"when": hook.When,
				})
			}
		}

		fmt.Println()
	}

	// Phase 6: Generate shell integration
	fmt.Println("🐚 Generating shell integration...")
	shellGen := shell.NewGenerator()

	// Detect shell type
	detectedShell := platform.DetectShell()
	shellType := shell.ParseShellType(detectedShell)

	// Generate to ~/.gdf/generated/init.sh
	shellPath := filepath.Join(gdfDir, "generated", "init.sh")
	if !applyDryRun {
		// Load global (unassociated) aliases
		ga, err := apps.LoadGlobalAliases(filepath.Join(gdfDir, "aliases.yaml"))
		if err != nil {
			fmt.Printf("⚠️  Warning: could not load global aliases: %v\n", err)
			ga = &apps.GlobalAliases{Aliases: make(map[string]string)}
		}

		if err := shellGen.Generate(resolvedApps, shellType, shellPath, ga.Aliases); err != nil {
			return fmt.Errorf("generating shell integration: %w", err)
		}
	}
	fmt.Println("   ✓ Shell integration updated")
	fmt.Println("   💡 Run: source ~/.gdf/generated/init.sh")
	logger.Log("shell_generate", shellPath, nil)

	// Phase 7: Save operation log
	if !applyDryRun {
		logPath, err := logger.Save(gdfDir)
		if err != nil {
			fmt.Printf("⚠️  Warning: could not save operation log: %v\n", err)
		} else if logPath != "" {
			fmt.Printf("\n📝 Operations logged to: %s\n", logPath)
		}
	}

	// Phase 8: Update state
	if !applyDryRun {
		st, err := state.LoadFromDir(gdfDir)
		if err != nil {
			fmt.Printf("⚠️  Warning: could not load state: %v\n", err)
		} else {
			// Add each profile to state
			for _, profile := range resolvedProfiles {
				st.AddProfile(profile.Name, profile.Apps)
			}

			// Save state
			if err := st.Save(filepath.Join(gdfDir, "state.yaml")); err != nil {
				fmt.Printf("⚠️  Warning: could not save state: %v\n", err)
			}
		}
	}

	fmt.Println("\n✅ Apply complete!")
	if applyDryRun {
		fmt.Println("   (No changes were made - this was a dry run)")
	}

	return nil
}
