package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track <path>",
	Short: "Track an existing dotfile",
	Long: `Move an existing file to the GDF repository and replace it with a symlink.
Automatically detects the app name or uses --app if provided.`,
	Args: cobra.ExactArgs(1),
	RunE: runTrack,
}

var targetApp string
var secretFlag bool

func init() {
	appCmd.AddCommand(trackCmd)
	trackCmd.Flags().StringVarP(&targetApp, "app", "a", "", "App bundle to add this file to")
	trackCmd.Flags().BoolVar(&secretFlag, "secret", false, "Mark file as secret (add to .gitignore)")
}

func runTrack(cmd *cobra.Command, args []string) error {
	path := args[0]
	expandedPath := platform.ExpandPath(path)

	// 1. Verify file exists
	info, err := os.Stat(expandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", expandedPath)
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("tracking directories is not yet supported")
	}

	// 2. Determine App Name
	appName := targetApp
	if appName == "" {
		appName = apps.DetectAppFromPath(expandedPath)
		fmt.Printf("Detected app: %s\n", appName)
	}

	// 3. Prepare config paths
	gdfDir := platform.ConfigDir()
	dotfilesDir := filepath.Join(gdfDir, "dotfiles")
	appPath := filepath.Join(gdfDir, "apps", appName+".yaml")

	// 4. Determine destination in repo
	relPath := filepath.Base(expandedPath)
	// If detected app implies a subdirectory (e.g. nvim/init.vim), handle it?
	// For now, simple structure: dotfiles/<app>/<filename>
	destDir := filepath.Join(dotfilesDir, appName)
	destPath := filepath.Join(destDir, relPath)
	filesSource := filepath.Join(appName, relPath)

	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("file already exists in repo: %s", destPath)
	}

	// 5. Move file
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	if err := moveFile(expandedPath, destPath); err != nil {
		return fmt.Errorf("moving file: %w", err)
	}

	// 6. Handle Secret (Update .gitignore BEFORE symlinking/committing)
	if secretFlag {
		gitignorePath := filepath.Join(gdfDir, ".gitignore")
		ignoreEntry := filepath.Join("dotfiles", filesSource)

		if err := addToGitignore(gitignorePath, ignoreEntry); err != nil {
			// Try to restore moved file
			_ = moveFile(destPath, expandedPath)
			return fmt.Errorf("updating .gitignore: %w", err)
		}
		fmt.Printf("✓ Added '%s' to .gitignore\n", ignoreEntry)
	}

	// 7. Symlink back
	if err := os.Symlink(destPath, expandedPath); err != nil {
		// Try to restore moved file if link fails
		_ = moveFile(destPath, expandedPath)
		return fmt.Errorf("creating symlink: %w", err)
	}

	// 8. Update App Bundle
	var bundle *apps.Bundle
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		// New app
		bundle = &apps.Bundle{
			Name:        appName,
			Description: fmt.Sprintf("App bundle for %s", appName),
		}
	} else {
		// Existing app
		bundle, err = apps.Load(appPath)
		if err != nil {
			return fmt.Errorf("loading app bundle: %w", err)
		}
	}

	// Calculate target for dotfile (relative to home if possible)
	home := platform.Detect().Home
	target := expandedPath
	if strings.HasPrefix(expandedPath, home) {
		target = "~" + strings.TrimPrefix(expandedPath, home)
	}

	// Add dotfile to bundle
	bundle.Dotfiles = append(bundle.Dotfiles, apps.Dotfile{
		Source: filesSource,
		Target: target,
		Secret: secretFlag,
	})

	if err := bundle.Save(appPath); err != nil {
		return fmt.Errorf("saving app bundle: %w", err)
	}

	fmt.Printf("✓ Tracked %s in app '%s'\n", path, appName)
	if secretFlag {
		fmt.Println("🔒 Marked as SECRET (gitignored)")
	}
	return nil
}

func addToGitignore(gitignorePath, entry string) error {
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Check if already exists (simple check)
	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), entry) {
		return nil
	}

	// Add newline if needed
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}

	_, err = f.WriteString(entry + "\n")
	return err
}

func moveFile(src, dst string) error {
	// Try rename first
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Fallback to copy+delete (cross-device)
	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
