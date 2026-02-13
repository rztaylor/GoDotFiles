package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	fmt.Println("  1. Add some apps:     gdf add kubectl")
	fmt.Println("  2. Track dotfiles:    gdf track ~/.gitconfig")
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

shell: zsh

conflict_resolution:
  aliases: last_wins    # last_wins, error, or prompt
  dotfiles: error       # error, backup_and_replace, or prompt

security:
  confirm_scripts: true
  log_scripts: true
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

	// Prompt user
	fmt.Print("Add shell integration to your RC file? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "n" || response == "no" {
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
	fmt.Println("\nTo activate in current session:")
	fmt.Println("  source ~/.gdf/generated/init.sh")
	fmt.Println("Or restart your shell.")

	return nil
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
