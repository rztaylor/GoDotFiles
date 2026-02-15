package cli

import (
	"fmt"
	"os"
	"path/filepath"

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

func init() {
	rootCmd.AddCommand(initCmd)
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
	content := `# GDF Configuration
# See: https://github.com/user/gdf/docs/reference/config.md
kind: Config/v1

shell: zsh

conflict_resolution:
  aliases: last_wins    # last_wins, error, or prompt
  dotfiles: error       # error, backup_and_replace, or prompt

security:
  confirm_scripts: true
  log_scripts: true

history:
  max_size_mb: 512
`
	path := filepath.Join(gdfDir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
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
