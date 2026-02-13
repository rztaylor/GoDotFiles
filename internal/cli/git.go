package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/git"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save [message]",
	Short: "Stage and commit all changes in the GDF repository",
	Long: `Save stages all changes in ~/.gdf and creates a commit.

With no arguments, uses default message "Update dotfiles":
  gdf save

With a message argument:
  gdf save "Added kubectl config"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSave,
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push commits to the remote repository",
	Long: `Push commits to the configured remote repository.

Requires a git remote to be configured:
  git -C ~/.gdf remote add origin <url>

Example:
  gdf push`,
	Args: cobra.NoArgs,
	RunE: runPush,
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull changes from the remote repository",
	Long: `Pull changes from the configured remote repository.

Requires a git remote to be configured:
  git -C ~/.gdf remote add origin <url>

Example:
  gdf pull`,
	Args: cobra.NoArgs,
	RunE: runPull,
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Full sync: pull, commit changes, and push",
	Long: `Sync performs a full synchronization workflow:
  1. Pull changes from remote
  2. Commit any local changes
  3. Push to remote

This is useful for keeping multiple machines in sync.

Requires a git remote to be configured:
  git -C ~/.gdf remote add origin <url>

Example:
  gdf sync`,
	Args: cobra.NoArgs,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(syncCmd)
}

func runSave(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()

	// Open repository
	repo, err := git.Open(gdfDir)
	if err != nil {
		return fmt.Errorf("opening repository: %w\nHint: Run 'gdf init' first", err)
	}

	// Check if there are changes
	hasChanges, err := repo.HasChanges()
	if err != nil {
		return fmt.Errorf("checking for changes: %w", err)
	}

	if !hasChanges {
		fmt.Println("No changes to save")
		return nil
	}

	// Stage all changes
	if err := repo.Add("."); err != nil {
		return fmt.Errorf("staging changes: %w", err)
	}

	// Determine commit message
	message := "Update dotfiles"
	if len(args) > 0 {
		message = args[0]
	}

	// Commit
	if err := repo.Commit(message); err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}

	fmt.Printf("✓ Saved changes: %s\n", message)
	return nil
}

func runPush(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()

	// Open repository
	repo, err := git.Open(gdfDir)
	if err != nil {
		return fmt.Errorf("opening repository: %w\nHint: Run 'gdf init' first", err)
	}

	// Check if remote exists
	if err := checkRemoteExists(gdfDir); err != nil {
		return err
	}

	// Push
	if err := repo.Push(); err != nil {
		// Check for common errors and provide helpful messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "no upstream branch") || strings.Contains(errMsg, "upstream") {
			return fmt.Errorf("no upstream branch configured\nHint: git -C %s push -u origin main", gdfDir)
		}
		return fmt.Errorf("pushing to remote: %w", err)
	}

	fmt.Println("✓ Pushed to remote")
	return nil
}

func runPull(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()

	// Open repository
	repo, err := git.Open(gdfDir)
	if err != nil {
		return fmt.Errorf("opening repository: %w\nHint: Run 'gdf init' first", err)
	}

	// Check if remote exists
	if err := checkRemoteExists(gdfDir); err != nil {
		return err
	}

	// Pull
	if err := repo.Pull(); err != nil {
		// Check for merge conflicts
		errMsg := err.Error()
		if strings.Contains(errMsg, "conflict") || strings.Contains(errMsg, "CONFLICT") {
			return fmt.Errorf("merge conflict detected\nHint: Resolve conflicts manually in %s", gdfDir)
		}
		return fmt.Errorf("pulling from remote: %w", err)
	}

	fmt.Println("✓ Pulled from remote")
	return nil
}

func runSync(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()

	// Open repository
	repo, err := git.Open(gdfDir)
	if err != nil {
		return fmt.Errorf("opening repository: %w\nHint: Run 'gdf init' first", err)
	}

	// Check if remote exists
	if err := checkRemoteExists(gdfDir); err != nil {
		return err
	}

	// Step 1: Pull
	fmt.Println("Pulling from remote...")
	if err := repo.Pull(); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "conflict") || strings.Contains(errMsg, "CONFLICT") {
			return fmt.Errorf("sync failed: merge conflict detected\nHint: Resolve conflicts manually in %s", gdfDir)
		}
		return fmt.Errorf("sync failed during pull: %w", err)
	}
	fmt.Println("✓ Pulled from remote")

	// Step 2: Check for local changes and commit if needed
	hasChanges, err := repo.HasChanges()
	if err != nil {
		return fmt.Errorf("sync failed: checking for changes: %w", err)
	}

	if hasChanges {
		fmt.Println("Committing local changes...")
		if err := repo.Add("."); err != nil {
			return fmt.Errorf("sync failed: staging changes: %w", err)
		}
		if err := repo.Commit("Update dotfiles"); err != nil {
			return fmt.Errorf("sync failed: creating commit: %w", err)
		}
		fmt.Println("✓ Committed local changes")
	} else {
		fmt.Println("No local changes to commit")
	}

	// Step 3: Push
	fmt.Println("Pushing to remote...")
	if err := repo.Push(); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "no upstream branch") || strings.Contains(errMsg, "upstream") {
			return fmt.Errorf("sync failed: no upstream branch configured\nHint: git -C %s push -u origin main", gdfDir)
		}
		return fmt.Errorf("sync failed during push: %w", err)
	}
	fmt.Println("✓ Pushed to remote")

	fmt.Println("\n✓ Sync complete")
	return nil
}

// checkRemoteExists verifies that a git remote is configured
func checkRemoteExists(repoPath string) error {
	cmd := exec.Command("git", "remote")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("checking for remote: %w", err)
	}

	remotes := strings.TrimSpace(string(output))
	if remotes == "" {
		return fmt.Errorf("no git remote configured\nHint: git -C %s remote add origin <url>", repoPath)
	}

	return nil
}
