package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/engine"
	"github.com/rztaylor/GoDotFiles/internal/library"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/shell"
	"github.com/rztaylor/GoDotFiles/internal/state"
	"github.com/rztaylor/GoDotFiles/internal/util"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [profiles...]",
	Short: "Apply one or more profiles",
	Long: `Apply profiles to the system.

This command will:
  1. Resolve profile dependencies (includes)
  2. Resolve app dependencies
  3. Install packages (if package manager available)
  4. Link dotfiles with conflict resolution
  5. Record apply hooks for package-less bundles
  6. Generate shell integration scripts (aliases, env, functions, init, completions)

All operations are logged to ~/.gdf/.operations/ for potential rollback.`,
	Example: `  gdf apply base work
  gdf apply
  gdf apply --dry-run sre
  gdf apply base`,
	Args: cobra.ArbitraryArgs,
	RunE: runApply,
}

var applyDryRun bool
var applyAllowRisky bool
var applyJSON bool
var applyRunHooks bool
var applyHookTimeout time.Duration
var applyRiskConfirmationPrompt = defaultRiskConfirmationPrompt

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "Show what would be done without making changes")
	applyCmd.Flags().BoolVar(&applyAllowRisky, "allow-risky", false, "Proceed without confirmation when high-risk scripts are detected")
	applyCmd.Flags().BoolVar(&applyJSON, "json", false, "Output dry-run plan as JSON")
	applyCmd.Flags().BoolVar(&applyRunHooks, "run-apply-hooks", false, "Execute hooks.apply commands (disabled by default)")
	applyCmd.Flags().DurationVar(&applyHookTimeout, "apply-hook-timeout", 30*time.Second, "Per-hook timeout when running hooks.apply")
}

func runApply(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()
	plat := platform.Detect()
	profileNames, err := resolveApplyProfileNames(args, gdfDir)
	if err != nil {
		return err
	}

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

	if applyJSON && !applyDryRun {
		return fmt.Errorf("--json is currently only supported with --dry-run")
	}
	if applyRunHooks && applyHookTimeout <= 0 {
		return fmt.Errorf("--apply-hook-timeout must be greater than 0")
	}
	if !applyDryRun {
		lock, err := acquireApplyLock(gdfDir)
		if err != nil {
			return err
		}
		defer func() {
			if releaseErr := lock.Release(); releaseErr != nil {
				fmt.Printf("⚠️  Warning: failed to release apply lock: %v\n", releaseErr)
			}
		}()
	}

	if applyDryRun {
		validation, err := runHealthValidateReport(gdfDir)
		if err != nil {
			return err
		}
		if validation.Errors > 0 {
			if applyJSON {
				if err := writeHealthJSON(cmd.OutOrStdout(), validation); err != nil {
					return err
				}
			}
			return withExitCode(fmt.Errorf("dry-run blocked by validation issues"), exitCodeHealthIssues)
		}
	}

	if applyJSON {
		return runApplyDryRunJSON(cmd, profileNames, gdfDir, plat, cfg)
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
	}

	// Phase 3: Load all app bundles (recursively)
	appsDir := filepath.Join(gdfDir, "apps")
	allBundles := make(map[string]*apps.Bundle)

	// Queue for recursive loading
	var queue []string
	for appName := range appNames {
		queue = append(queue, appName)
	}

	libMgr := library.New() // Initialize library manager

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]

		if _, exists := allBundles[name]; exists {
			continue
		}

		// Try loading from local apps directory first
		appPath := filepath.Join(appsDir, name+".yaml")
		bundle, err := apps.Load(appPath)
		if err != nil {
			// If not found locally, try the library
			if os.IsNotExist(err) {
				fmt.Printf("   ℹ️  App '%s' not found locally, checking library...\n", name)
				recipe, libErr := libMgr.Get(name)
				if libErr == nil {
					// Found in library, instantiate in-memory
					bundle = recipe.ToBundle()
					fmt.Printf("   ✨ Resolved '%s' from library (in-memory)\n", name)
				} else {
					fmt.Printf("⚠️  Warning: skipping app '%s': %v\n", name, err)
					continue
				}
			} else {
				fmt.Printf("⚠️  Warning: error loading app '%s': %v\n", name, err)
				continue
			}
		}

		allBundles[name] = bundle

		// Add dependencies to queue
		if len(bundle.Dependencies) > 0 {
			queue = append(queue, bundle.Dependencies...)
		}
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

	// Security scan before any mutating operations.
	findings := engine.DetectHighRiskConfigurations(resolvedApps)
	if !applyRunHooks {
		findings = filterRiskFindingsForPolicy(findings)
	}
	if len(findings) > 0 {
		fmt.Println("\n⚠️  High-risk commands detected:")
		for i, f := range findings {
			fmt.Printf("   %d. app=%s, location=%s\n", i+1, f.App, f.Location)
			fmt.Printf("      reason: %s\n", f.Reason)
			fmt.Printf("      command: %s\n", f.Command)
		}

		confirmScripts := true
		if cfg.Security != nil {
			confirmScripts = cfg.Security.ConfirmScriptsDefault()
		}
		if !applyAllowRisky && confirmScripts {
			confirmed, err := applyRiskConfirmationPrompt(findings)
			if err != nil {
				return err
			}
			if !confirmed {
				return fmt.Errorf("aborted due to high-risk configuration")
			}
		}
	}

	// Phase 5: Apply each app
	conflictStrategy := "error"
	if cfg.ConflictResolution != nil {
		conflictStrategy = cfg.ConflictResolution.DotfilesDefault()
	}
	linker := engine.NewLinker(conflictStrategy)
	linker.SetHistoryManager(engine.NewHistoryManager(gdfDir, cfg.History.MaxSizeMBDefault()))

	for _, bundle := range resolvedApps {
		fmt.Printf("📦 Processing app: %s\n", bundle.Name)

		// 5a. Install package (if defined)
		if bundle.Package != nil {
			plan := resolvePackageManagerPlan(bundle.Package, plat, cfg)
			if plan == nil {
				fmt.Printf("      ⚠️  App '%s' has no supported package manager configuration for this system. Skipping package install.\n", bundle.Name)
				goto SkipPackageInstall
			}

			selected := plan.Selected
			if selected.Name == "custom" {
				fmt.Println("   Package: custom install script")
				fmt.Println("      ⏭️  Skipping custom script execution during apply")
				logger.Log("package_install_skipped", "custom_script", map[string]string{
					"manager": "custom",
					"app":     bundle.Name,
					"reason":  "custom_script_not_executed_in_apply",
				})
				goto SkipPackageInstall
			}

			fmt.Printf("   Package: %s (via %s)\n", selected.PackageName, selected.Name)

			if selected.Name != "none" {
				alreadyInstalled := false
				detectedBy := ""
				if !applyDryRun {
					for _, probe := range plan.Probes {
						installed, checkErr := probe.Manager.IsInstalled(probe.PackageName)
						if checkErr != nil {
							fmt.Printf("      ⚠️  Could not verify install status for '%s' via %s: %v\n", probe.PackageName, probe.Name, checkErr)
							continue
						}
						if installed {
							alreadyInstalled = true
							detectedBy = probe.Name
							break
						}
					}
					if alreadyInstalled {
						if detectedBy != "" && detectedBy != selected.Name {
							fmt.Printf("      ⏭️  Skipping install (already installed via %s)\n", detectedBy)
						} else {
							fmt.Printf("      ⏭️  Skipping install (already installed)\n")
						}
					}
				}

				if alreadyInstalled {
					logger.Log("package_install_skipped", selected.PackageName, map[string]string{
						"manager":          selected.Name,
						"app":              bundle.Name,
						"reason":           "already_installed",
						"detected_manager": detectedBy,
					})
				} else {
					if !applyDryRun {
						if err := selected.Manager.Install(selected.PackageName); err != nil {
							return fmt.Errorf("installing package %s: %w", selected.PackageName, err)
						}
					}
					logger.Log("package_install", selected.PackageName, map[string]string{
						"manager": selected.Name,
						"app":     bundle.Name,
					})
				}
			} else {
				fmt.Printf("      ⏭️  Skipping (no package manager)\n")
			}
		}

	SkipPackageInstall:
		// 5b. Link dotfiles
		if len(bundle.Dotfiles) > 0 {
			fmt.Printf("   Dotfiles: %d file(s)\n", len(bundle.Dotfiles))
			for _, dotfile := range bundle.Dotfiles {
				if dotfile.When != "" {
					match, err := config.EvaluateCondition(dotfile.When, plat)
					if err != nil {
						return fmt.Errorf("evaluating condition for dotfile %s in app %s: %w", dotfile.Source, bundle.Name, err)
					}
					if !match {
						fmt.Printf("      ⏭️  skip %s (condition: %s)\n", dotfile.Source, dotfile.When)
						continue
					}
				}

				effectiveTarget := dotfile.EffectiveTarget(plat.OS)
				if effectiveTarget == "" {
					return fmt.Errorf("dotfile %s in app %s has no target for os %s", dotfile.Source, bundle.Name, plat.OS)
				}

				dotfileToLink := dotfile
				dotfileToLink.Target = effectiveTarget

				if !applyDryRun {
					if err := linker.Link(dotfileToLink, gdfDir); err != nil {
						return fmt.Errorf("linking %s: %w", dotfile.Source, err)
					}
				}
				fmt.Printf("      ✓ %s → %s\n", effectiveTarget, dotfile.Source)
				details := map[string]string{
					"source": dotfile.Source,
					"app":    bundle.Name,
					// Absolute source allows safer rollback checks.
					"source_abs": filepath.Join(gdfDir, "dotfiles", dotfile.Source),
				}
				if snap := linker.ConsumeConflictSnapshot(platform.ExpandPath(effectiveTarget)); snap != nil {
					details["snapshot_id"] = snap.ID
					details["snapshot_path"] = snap.Path
					details["snapshot_kind"] = snap.Kind
					details["snapshot_link_target"] = snap.LinkTarget
					details["snapshot_mode"] = fmt.Sprintf("%#o", uint32(snap.Mode.Perm()))
					details["snapshot_checksum"] = snap.Checksum
					details["snapshot_size_bytes"] = fmt.Sprintf("%d", snap.SizeBytes)
					details["snapshot_captured_at"] = snap.CapturedAt.Format("2006-01-02T15:04:05.999999999Z07:00")
				}
				logger.Log("link", effectiveTarget, details)
			}
		}

		// 5c. Run apply hooks (for package-less bundles)
		if bundle.Hooks != nil && len(bundle.Hooks.Apply) > 0 {
			fmt.Printf("   Apply hooks: %d command(s)\n", len(bundle.Hooks.Apply))
			for _, hook := range bundle.Hooks.Apply {
				fmt.Printf("      • %s\n", hook.Run)
				if hook.When != "" {
					match, err := config.EvaluateCondition(hook.When, plat)
					if err != nil {
						return fmt.Errorf("evaluating apply hook condition for app %s: %w", bundle.Name, err)
					}
					if !match {
						fmt.Printf("      ⏭️  Skipping hook (condition: %s)\n", hook.When)
						logger.Log("hook_skip", hook.Run, map[string]string{
							"type":   "apply",
							"app":    bundle.Name,
							"when":   hook.When,
							"reason": "condition_not_met",
						})
						continue
					}
				}
				if applyDryRun {
					logger.Log("hook_run", hook.Run, map[string]string{
						"type":    "apply",
						"app":     bundle.Name,
						"when":    hook.When,
						"dry_run": "true",
					})
					continue
				}

				if !applyRunHooks {
					fmt.Println("      ⏭️  Skipping (hooks.apply execution is disabled; use --run-apply-hooks)")
					logger.Log("hook_skip", hook.Run, map[string]string{
						"type":   "apply",
						"app":    bundle.Name,
						"when":   hook.When,
						"reason": "not_opted_in",
					})
					continue
				}
				if err := executeApplyHook(hook.Run, applyHookTimeout); err != nil {
					return fmt.Errorf("running apply hook for app %s: %w", bundle.Name, err)
				}
				logger.Log("hook_run", hook.Run, map[string]string{
					"type":    "apply",
					"app":     bundle.Name,
					"when":    hook.When,
					"timeout": applyHookTimeout.String(),
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
		compCount, compWarnings, err := generateManagedCompletionFiles(resolvedApps, gdfDir)
		if err != nil {
			return fmt.Errorf("generating managed shell completion files: %w", err)
		}
		if compCount > 0 {
			fmt.Printf("   ✓ Managed shell completions updated (%d)\n", compCount)
		}
		for _, warning := range compWarnings {
			fmt.Printf("   ⚠️  %s\n", warning)
		}

		// Load global (unassociated) aliases
		ga, err := apps.LoadGlobalAliases(filepath.Join(gdfDir, "aliases.yaml"))
		if err != nil {
			fmt.Printf("⚠️  Warning: could not load global aliases: %v\n", err)
			ga = &apps.GlobalAliases{Aliases: make(map[string]string)}
		}

		opts := shell.GenerateOptions{
			EnableAutoReload:          cfg.ShellIntegration.AutoReloadEnabledDefault(),
			DisableCompletionCommands: true,
		}
		if err := shellGen.GenerateWithOptions(resolvedApps, shellType, shellPath, ga.Aliases, opts); err != nil {
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

func resolveApplyProfileNames(requested []string, gdfDir string) ([]string, error) {
	if len(requested) > 0 {
		return requested, nil
	}

	st, err := state.LoadFromDir(gdfDir)
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}
	if len(st.AppliedProfiles) > 0 {
		names := make([]string, 0, len(st.AppliedProfiles))
		for _, profile := range st.AppliedProfiles {
			if profile.Name == "" {
				continue
			}
			names = append(names, profile.Name)
		}
		if len(names) > 0 {
			fmt.Printf("No profile arguments provided. Reapplying profiles from local state: %s\n", strings.Join(names, ", "))
			return names, nil
		}
	}

	selected, err := resolveProfileSelection(gdfDir, "")
	if err != nil {
		return nil, err
	}
	fmt.Printf("No profile arguments provided. Applying selected profile: %s\n", selected)
	return []string{selected}, nil
}

func defaultRiskConfirmationPrompt(findings []engine.RiskFinding) (bool, error) {
	if globalNonInteractive && !globalYes {
		return false, withExitCode(
			fmt.Errorf("high-risk configuration detected in non-interactive mode; re-run with --allow-risky or --yes to proceed"),
			exitCodeNonInteractiveStop,
		)
	}

	return confirmPrompt("\nProceed despite these high-risk commands? [y/N]: ")
}

func generateManagedCompletionFiles(bundles []*apps.Bundle, gdfDir string) (int, []string, error) {
	baseDir := filepath.Join(gdfDir, "generated", "completions")
	bashDir := filepath.Join(baseDir, "bash")
	zshDir := filepath.Join(baseDir, "zsh")

	for _, dir := range []string{bashDir, zshDir} {
		if err := os.RemoveAll(dir); err != nil {
			return 0, nil, fmt.Errorf("clearing completion dir %s: %w", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return 0, nil, fmt.Errorf("creating completion dir %s: %w", dir, err)
		}
	}

	bundleMap := make(map[string]*apps.Bundle, len(bundles))
	names := make([]string, 0, len(bundles))
	for _, b := range bundles {
		bundleMap[b.Name] = b
		names = append(names, b.Name)
	}
	sort.Strings(names)

	var count int
	var warnings []string
	for _, name := range names {
		bundle := bundleMap[name]
		if bundle.Shell == nil || bundle.Shell.Completions == nil {
			continue
		}

		if cmd := strings.TrimSpace(bundle.Shell.Completions.Bash); cmd != "" {
			output, err := runCompletionCommand(cmd)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Skipping bash completion for app '%s': %v", bundle.Name, err))
			} else {
				path := filepath.Join(bashDir, completionFileName(bundle.Name))
				if err := util.WriteFileAtomic(path, output, 0644); err != nil {
					return 0, nil, fmt.Errorf("writing bash completion for app %s: %w", bundle.Name, err)
				}
				count++
			}
		}

		if cmd := strings.TrimSpace(bundle.Shell.Completions.Zsh); cmd != "" {
			output, err := runCompletionCommand(cmd)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Skipping zsh completion for app '%s': %v", bundle.Name, err))
			} else {
				path := filepath.Join(zshDir, completionFileName(bundle.Name))
				if err := util.WriteFileAtomic(path, output, 0644); err != nil {
					return 0, nil, fmt.Errorf("writing zsh completion for app %s: %w", bundle.Name, err)
				}
				count++
			}
		}
	}

	return count, warnings, nil
}

func runCompletionCommand(command string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("command failed (%s): %s", command, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("running command (%s): %w", command, err)
	}
	if len(output) == 0 {
		return nil, fmt.Errorf("command produced no output (%s)", command)
	}
	return output, nil
}

func completionFileName(appName string) string {
	if appName == "" {
		return "app.sh"
	}
	var b strings.Builder
	for _, r := range appName {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('-')
	}
	name := b.String()
	if name == "" {
		name = "app"
	}
	return name + ".sh"
}

func filterRiskFindingsForPolicy(findings []engine.RiskFinding) []engine.RiskFinding {
	filtered := make([]engine.RiskFinding, 0, len(findings))
	for _, finding := range findings {
		if finding.Location == "hooks.apply.run" {
			continue
		}
		filtered = append(filtered, finding)
	}
	return filtered
}

func executeApplyHook(command string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("hook timed out after %s", timeout)
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed != "" {
		return fmt.Errorf("hook command failed: %w: %s", err, trimmed)
	}
	return fmt.Errorf("hook command failed: %w", err)
}

type applyLock struct {
	path string
}

func acquireApplyLock(gdfDir string) (*applyLock, error) {
	lockDir := filepath.Join(gdfDir, ".locks")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return nil, fmt.Errorf("creating apply lock directory: %w", err)
	}

	lockPath := filepath.Join(lockDir, "apply.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("another apply operation is already in progress (lock file: %s)", lockPath)
		}
		return nil, fmt.Errorf("acquiring apply lock: %w", err)
	}
	if _, writeErr := fmt.Fprintf(f, "pid=%d\ncreated=%s\n", os.Getpid(), time.Now().Format(time.RFC3339Nano)); writeErr != nil {
		_ = f.Close()
		_ = os.Remove(lockPath)
		return nil, fmt.Errorf("writing apply lock metadata: %w", writeErr)
	}
	if closeErr := f.Close(); closeErr != nil {
		_ = os.Remove(lockPath)
		return nil, fmt.Errorf("closing apply lock file: %w", closeErr)
	}
	return &applyLock{path: lockPath}, nil
}

func (l *applyLock) Release() error {
	if l == nil || l.path == "" {
		return nil
	}
	err := os.Remove(l.path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return err
}
