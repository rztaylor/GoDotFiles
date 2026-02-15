package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/git"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/shell"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [url]",
	Short: "Initialize a new GDF repository or clone an existing one",
	Long: `Initialize sets up GDF in ~/.gdf.

With no arguments, creates a new repository:
  gdf init

With a URL, clones an existing dotfiles repository:
  gdf init git@github.com:user/dotfiles.git`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

var initSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run first-run interactive setup",
	Long: `Run a guided first-run setup for profile/app bootstrap.

Creates ~/.gdf if needed, helps select a profile and starter apps, and optionally outputs JSON summary.`,
	RunE: runInitSetup,
}

var (
	setupProfile string
	setupApps    string
	setupJSON    bool
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.AddCommand(initSetupCmd)

	initSetupCmd.Flags().StringVarP(&setupProfile, "profile", "p", "", "Profile name to bootstrap (default: prompt or 'default')")
	initSetupCmd.Flags().StringVar(&setupApps, "apps", "", "Comma-separated starter apps to add to the selected profile")
	initSetupCmd.Flags().BoolVar(&setupJSON, "json", false, "Output setup summary as JSON")
}

func runInit(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()

	// Check if already initialized
	if git.IsRepository(gdfDir) {
		return fmt.Errorf("GDF is already initialized at %s", gdfDir)
	}

	if len(args) == 1 {
		// Clone existing repo
		return cloneRepo(args[0], gdfDir)
	}

	// Create new repo
	return createNewRepo(gdfDir)
}

type initSetupSummary struct {
	Initialized bool     `json:"initialized"`
	Profile     string   `json:"profile"`
	Apps        []string `json:"apps"`
}

func runInitSetup(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()
	initialized := git.IsRepository(gdfDir)

	if !initialized {
		if err := createNewRepo(gdfDir); err != nil {
			return err
		}
		initialized = true
	} else {
		fmt.Printf("Using existing GDF repository at %s\n", gdfDir)
	}

	profileName := strings.TrimSpace(setupProfile)
	if profileName == "" {
		profileName = "default"
		if !globalNonInteractive {
			input, err := readInteractiveLine("Profile name to bootstrap [default]: ")
			if err != nil {
				return err
			}
			input = strings.TrimSpace(input)
			if input != "" {
				profileName = input
			}
		}
	}

	appNames := parseCSVAppNames(setupApps)
	if len(appNames) == 0 && !globalNonInteractive {
		input, err := readInteractiveLine("Starter apps (comma-separated, optional): ")
		if err != nil {
			return err
		}
		appNames = parseCSVAppNames(input)
	}

	for _, appName := range appNames {
		if err := addAppToProfile(gdfDir, profileName, appName); err != nil {
			return err
		}
		appPath := filepath.Join(gdfDir, "apps", appName+".yaml")
		if _, err := os.Stat(appPath); os.IsNotExist(err) {
			if err := createAppSkeleton(AppName(appName), appPath); err != nil {
				return fmt.Errorf("creating app skeleton for %s: %w", appName, err)
			}
		}
	}

	summary := initSetupSummary{
		Initialized: initialized,
		Profile:     profileName,
		Apps:        appNames,
	}
	if setupJSON {
		data, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling setup summary: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println("✓ Setup complete")
	fmt.Printf("Profile: %s\n", profileName)
	if len(appNames) == 0 {
		fmt.Println("Starter apps: (none)")
	} else {
		fmt.Printf("Starter apps: %s\n", strings.Join(appNames, ", "))
	}
	fmt.Println("Next: run `gdf apply` when ready.")
	return nil
}

func parseCSVAppNames(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		name := AppName(strings.TrimSpace(part))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func createNewRepo(gdfDir string) error {
	fmt.Printf("Initializing new GDF repository at %s\n", gdfDir)

	// Initialize git repo
	repo, err := git.Init(gdfDir)
	if err != nil {
		return fmt.Errorf("initializing repository: %w", err)
	}

	// Create directory structure
	dirs := []string{
		"apps",
		"profiles",
		"dotfiles",
	}
	for _, dir := range dirs {
		path := filepath.Join(gdfDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("creating %s directory: %w", dir, err)
		}
	}

	// Create .gitignore
	if err := createGitignore(gdfDir); err != nil {
		return err
	}

	// Create initial config.yaml
	if err := createInitialConfig(gdfDir); err != nil {
		return err
	}

	// Create default profile
	if err := createDefaultProfile(gdfDir); err != nil {
		return err
	}

	// Initial commit
	if err := repo.Add("."); err != nil {
		return fmt.Errorf("staging files: %w", err)
	}
	if err := repo.Commit("Initial GDF setup"); err != nil {
		return fmt.Errorf("creating initial commit: %w", err)
	}

	if err := ensureGeneratedInitScript(gdfDir); err != nil {
		return err
	}

	fmt.Println("✓ Repository initialized")

	// Offer to inject shell integration
	if err := offerShellInjection(); err != nil {
		// Warn but don't fail the init
		fmt.Printf("⚠ Could not set up shell integration: %v\n", err)
		fmt.Println("  You can add it manually later.")
	}

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add some apps:     gdf app add kubectl")
	fmt.Println("  2. Track dotfiles:    gdf app track ~/.gitconfig")
	fmt.Println("  3. Set up a remote:   git -C ~/.gdf remote add origin <url>")

	return nil
}

func cloneRepo(url, gdfDir string) error {
	fmt.Printf("Cloning repository from %s\n", url)

	_, err := git.Clone(url, gdfDir)
	if err != nil {
		return fmt.Errorf("cloning repository: %w", err)
	}

	fmt.Println("✓ Repository cloned")

	if err := ensureGeneratedInitScript(gdfDir); err != nil {
		return err
	}

	// Offer to inject shell integration
	if err := offerShellInjection(); err != nil {
		// Warn but don't fail
		fmt.Printf("⚠ Could not set up shell integration: %v\n", err)
		fmt.Println("  You can add it manually later.")
	}

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Apply your profile: gdf apply")

	return nil
}

func ensureGeneratedInitScript(gdfDir string) error {
	generatedDir := filepath.Join(gdfDir, "generated")
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		return fmt.Errorf("creating generated directory: %w", err)
	}

	scriptPath := filepath.Join(generatedDir, "init.sh")
	if _, err := os.Stat(scriptPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking generated init script: %w", err)
	}

	content := `#!/usr/bin/env sh
# Placeholder created by gdf init.
# Run 'gdf apply' to generate shell integration.
`
	if err := os.WriteFile(scriptPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("creating generated init script: %w", err)
	}

	return nil
}

func createGitignore(gdfDir string) error {
	content := `# GDF local-only files
state.yaml
.operations/
.history/

# Editor files
*.swp
*~
.DS_Store
`
	path := filepath.Join(gdfDir, ".gitignore")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("creating .gitignore: %w", err)
	}
	return nil
}

func createInitialConfig(gdfDir string) error {
	path := filepath.Join(gdfDir, "config.yaml")
	if err := config.WriteDefaultConfig(path, platform.DetectShell()); err != nil {
		return fmt.Errorf("creating config.yaml: %w", err)
	}
	return nil
}

func createDefaultProfile(gdfDir string) error {
	profileDir := filepath.Join(gdfDir, "profiles", "default")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("creating default profile directory: %w", err)
	}

	content := `# Default Profile
kind: Profile/v1
name: default
description: Default profile with common apps

# Add apps to this profile:
# apps:
#   - git
#   - vim
`
	path := filepath.Join(profileDir, "profile.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("creating profile.yaml: %w", err)
	}
	return nil
}

// offerShellInjection prompts the user to add shell integration to their RC file.
func offerShellInjection() error {
	detectedShell := platform.DetectShell()
	shellType := shell.ParseShellType(detectedShell)

	if shellType == shell.Unknown {
		fmt.Println("\n⚠ Could not detect your shell")
		fmt.Println("  To enable shell integration, add to your shell config:")
		fmt.Println("  [ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh")
		return nil
	}

	fmt.Printf("\nDetected shell: %s\n", detectedShell)

	if globalNonInteractive {
		fmt.Println("Skipping interactive shell setup in non-interactive mode.")
		fmt.Println("To add it manually, add this to your shell config:")
		fmt.Println("  [ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh")
		return nil
	}

	// Prompt user
	confirmInjection, err := confirmPromptDefaultYes("Add shell integration to your RC file? [Y/n]: ")
	if err != nil {
		return err
	}

	if !confirmInjection {
		fmt.Println("\nSkipped shell integration.")
		fmt.Println("To add it manually, add this to your shell config:")
		fmt.Println("  [ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh")
		return nil
	}

	// Inject
	injector := shell.NewInjector()
	if err := injector.InjectSourceLine(shellType); err != nil {
		return err
	}

	fmt.Println("✓ Added shell integration")
	fmt.Printf("  Source line added to %s\n", getRCFileName(shellType))

	if shellType == shell.Bash || shellType == shell.Zsh {
		fmt.Println()
		enableAutoReload, err := confirmPromptDefaultYes("Enable event-based shell auto-reload on prompt? [Y/n]: ")
		if err != nil {
			return err
		}
		if enableAutoReload {
			cfgPath := filepath.Join(platform.ConfigDir(), "config.yaml")
			if err := setAutoReloadEnabled(cfgPath, true); err != nil {
				return fmt.Errorf("enabling shell auto-reload: %w", err)
			}
			fmt.Println("✓ Enabled shell auto-reload in config")
		} else {
			fmt.Println("Skipped auto-reload setup.")
		}

		fmt.Println()
		installCompletion, err := confirmPromptDefaultYes("Install gdf shell completion now? [Y/n]: ")
		if err != nil {
			return err
		}
		if installCompletion {
			completionPath, err := installShellCompletion(shellType)
			if err != nil {
				fmt.Printf("⚠ Could not install shell completion automatically: %v\n", err)
				printManualCompletionInstructions(shellType)
			} else {
				fmt.Printf("✓ Installed shell completion: %s\n", completionPath)
			}
		} else {
			fmt.Println("Skipped shell completion setup.")
			printManualCompletionInstructions(shellType)
		}
	}

	fmt.Println("\nTo activate in current session:")
	fmt.Println("  source ~/.gdf/generated/init.sh")
	fmt.Println("Or restart your shell.")

	return nil
}

func setAutoReloadEnabled(configPath string, enabled bool) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}
	if cfg.ShellIntegration == nil {
		cfg.ShellIntegration = &config.ShellIntegrationConfig{}
	}
	cfg.ShellIntegration.AutoReloadEnabled = &enabled
	return cfg.Save(configPath)
}

func installShellCompletion(shellType shell.ShellType) (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME is not set")
	}

	var completionPath string
	switch shellType {
	case shell.Bash:
		completionPath = filepath.Join(home, ".local", "share", "bash-completion", "completions", "gdf")
	case shell.Zsh:
		completionPath = filepath.Join(home, ".zfunc", "_gdf")
	default:
		return "", fmt.Errorf("unsupported shell for completion install: %s", shellType)
	}

	if err := os.MkdirAll(filepath.Dir(completionPath), 0755); err != nil {
		return "", fmt.Errorf("creating completion directory: %w", err)
	}

	f, err := os.Create(completionPath)
	if err != nil {
		return "", fmt.Errorf("creating completion file: %w", err)
	}
	defer f.Close()

	switch shellType {
	case shell.Bash:
		if err := rootCmd.GenBashCompletionV2(f, true); err != nil {
			return "", fmt.Errorf("generating bash completion: %w", err)
		}
	case shell.Zsh:
		if err := rootCmd.GenZshCompletion(f); err != nil {
			return "", fmt.Errorf("generating zsh completion: %w", err)
		}
	}

	return completionPath, nil
}

func printManualCompletionInstructions(shellType shell.ShellType) {
	switch shellType {
	case shell.Bash:
		fmt.Println("To install manually:")
		fmt.Println("  gdf shell completion bash > ~/.local/share/bash-completion/completions/gdf")
	case shell.Zsh:
		fmt.Println("To install manually:")
		fmt.Println("  mkdir -p ~/.zfunc")
		fmt.Println("  gdf shell completion zsh > ~/.zfunc/_gdf")
	default:
		fmt.Println("Generate completion manually with:")
		fmt.Println("  gdf shell completion bash")
		fmt.Println("  gdf shell completion zsh")
	}
}

// getRCFileName returns the RC file name for display purposes.
func getRCFileName(shellType shell.ShellType) string {
	switch shellType {
	case shell.Bash:
		return "~/.bashrc"
	case shell.Zsh:
		return "~/.zshrc"
	default:
		return "RC file"
	}
}
