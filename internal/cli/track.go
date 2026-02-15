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
var trackInteractive bool

type trackFileOptions struct {
	AppName     string
	Secret      bool
	Interactive bool
	Audit       *decisionAudit
}

type trackFileResult struct {
	Skipped bool
	Reason  string
}

func init() {
	appCmd.AddCommand(trackCmd)
	trackCmd.Flags().StringVarP(&targetApp, "app", "a", "", "App bundle to add this file to")
	trackCmd.Flags().BoolVar(&secretFlag, "secret", false, "Mark file as secret (add to .gitignore)")
	trackCmd.Flags().BoolVar(&trackInteractive, "interactive", false, "Preview and resolve path/target conflicts interactively")
}

func runTrack(cmd *cobra.Command, args []string) error {
	audit := newDecisionAudit("gdf app track", false)
	result, err := trackFile(args[0], trackFileOptions{
		AppName:     targetApp,
		Secret:      secretFlag,
		Interactive: trackInteractive,
		Audit:       audit,
	})
	if err != nil {
		return err
	}
	if result.Skipped {
		fmt.Printf("Skipped tracking %s (%s)\n", args[0], result.Reason)
	}
	if logPath, err := audit.Save(platform.ConfigDir()); err != nil {
		return err
	} else if logPath != "" {
		fmt.Printf("Logged conflict decisions: %s\n", logPath)
	}
	return nil
}

func trackFile(path string, opts trackFileOptions) (*trackFileResult, error) {
	expandedPath := platform.ExpandPath(path)

	// 1. Verify file exists
	info, err := os.Stat(expandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", expandedPath)
		}
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("tracking directories is not yet supported")
	}

	// 2. Determine App Name
	appName := opts.AppName
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
	destDir := filepath.Join(dotfilesDir, appName)
	destPath := filepath.Join(destDir, relPath)
	filesSource := filepath.Join(appName, relPath)

	if _, err := os.Stat(destPath); err == nil {
		if !opts.Interactive {
			return nil, fmt.Errorf("file already exists in repo: %s", destPath)
		}
		decision, err := chooseTrackConflictDecision(path, "repo path already exists", []string{"skip", "overwrite", "rename"})
		if err != nil {
			return nil, err
		}
		if opts.Audit != nil {
			opts.Audit.Record(path, "repo-path-exists", decision)
		}
		switch decision {
		case "skip":
			return &trackFileResult{Skipped: true, Reason: "repo path already exists"}, nil
		case "overwrite":
			if err := os.Remove(destPath); err != nil {
				return nil, fmt.Errorf("removing existing destination file: %w", err)
			}
		case "rename":
			relPath = uniqueTrackedFilename(destDir, relPath)
			destPath = filepath.Join(destDir, relPath)
			filesSource = filepath.Join(appName, relPath)
		}
	}

	// 5. Move file
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("creating destination directory: %w", err)
	}

	if err := moveFile(expandedPath, destPath); err != nil {
		return nil, fmt.Errorf("moving file: %w", err)
	}

	// 6. Handle Secret (Update .gitignore BEFORE symlinking/committing)
	if opts.Secret {
		gitignorePath := filepath.Join(gdfDir, ".gitignore")
		ignoreEntry := filepath.Join("dotfiles", filesSource)

		if err := addToGitignore(gitignorePath, ignoreEntry); err != nil {
			// Try to restore moved file
			_ = moveFile(destPath, expandedPath)
			return nil, fmt.Errorf("updating .gitignore: %w", err)
		}
		fmt.Printf("✓ Added '%s' to .gitignore\n", ignoreEntry)
	}

	// 7. Symlink back
	if err := os.Symlink(destPath, expandedPath); err != nil {
		// Try to restore moved file if link fails
		_ = moveFile(destPath, expandedPath)
		return nil, fmt.Errorf("creating symlink: %w", err)
	}

	// 8. Update App Bundle
	var bundle *apps.Bundle
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		bundle = &apps.Bundle{
			Name:        appName,
			Description: fmt.Sprintf("App bundle for %s", appName),
		}
	} else {
		bundle, err = apps.Load(appPath)
		if err != nil {
			return nil, fmt.Errorf("loading app bundle: %w", err)
		}
	}

	home := platform.Detect().Home
	target := expandedPath
	if strings.HasPrefix(expandedPath, home) {
		target = "~" + strings.TrimPrefix(expandedPath, home)
	}

	if idx := findDotfileTargetConflict(bundle.Dotfiles, target); idx >= 0 {
		if !opts.Interactive {
			return nil, fmt.Errorf("dotfile target already tracked in app '%s': %s", appName, target)
		}
		decision, err := chooseTrackConflictDecision(path, "target already tracked in app bundle", []string{"skip", "replace"})
		if err != nil {
			return nil, err
		}
		if opts.Audit != nil {
			opts.Audit.Record(path, "target-conflict", decision)
		}
		if decision == "skip" {
			return &trackFileResult{Skipped: true, Reason: "target already tracked"}, nil
		}
		bundle.Dotfiles[idx] = apps.Dotfile{Source: filesSource, Target: target, Secret: opts.Secret}
	} else {
		bundle.Dotfiles = append(bundle.Dotfiles, apps.Dotfile{
			Source: filesSource,
			Target: target,
			Secret: opts.Secret,
		})
	}

	if err := bundle.Save(appPath); err != nil {
		return nil, fmt.Errorf("saving app bundle: %w", err)
	}

	fmt.Printf("✓ Tracked %s in app '%s'\n", path, appName)
	if opts.Secret {
		fmt.Println("Marked as SECRET (gitignored)")
	}

	return &trackFileResult{}, nil
}

func chooseTrackConflictDecision(subject, conflict string, options []string) (string, error) {
	if globalNonInteractive {
		return "", withExitCode(fmt.Errorf("unresolved conflict for %s: %s", subject, conflict), exitCodeNonInteractiveStop)
	}

	fmt.Printf("\nConflict for %s\n", subject)
	fmt.Printf("  %s\n", conflict)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}

	input, err := readInteractiveLine("Choose an option [1]: ")
	if err != nil {
		return "", err
	}
	choice := strings.TrimSpace(input)
	if choice == "" {
		return options[0], nil
	}
	for i, opt := range options {
		if choice == fmt.Sprintf("%d", i+1) || strings.EqualFold(choice, opt) {
			return opt, nil
		}
	}
	return "", fmt.Errorf("invalid choice %q", choice)
}

func uniqueTrackedFilename(destDir, base string) string {
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s-imported-%d%s", name, i, ext)
		if _, err := os.Stat(filepath.Join(destDir, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
	return fmt.Sprintf("%s-imported%s", name, ext)
}

func findDotfileTargetConflict(dotfiles []apps.Dotfile, target string) int {
	for i, dot := range dotfiles {
		if dot.Target == target {
			return i
		}
	}
	return -1
}

func addToGitignore(gitignorePath, entry string) error {
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), entry) {
		return nil
	}

	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}

	_, err = f.WriteString(entry + "\n")
	return err
}

func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
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
