package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

// Restore unlinks the dotfile and copies the source file back to the target location.
// Returns nil if target is not a symlink or does not point to the expected source.
func (l *Linker) Restore(dotfile apps.Dotfile, gdfDir string) error {
	sourcePath := filepath.Join(gdfDir, "dotfiles", dotfile.Source)
	targetPath := platform.ExpandPath(dotfile.Target)

	// Check if target exists and is a symlink
	info, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to restore
		}
		return err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		// Not a symlink, leave it alone
		return nil
	}

	// Check if symlink points to our source
	dest, err := os.Readlink(targetPath)
	if err != nil {
		return fmt.Errorf("reading symlink: %w", err)
	}

	// Resolve paths for comparison
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("resolving source path: %w", err)
	}
	absDest, err := filepath.Abs(dest)
	if err != nil {

		if !filepath.IsAbs(dest) {
			absDest = filepath.Join(filepath.Dir(targetPath), dest)
			absDest, err = filepath.Abs(absDest)
			if err != nil {
				return fmt.Errorf("resolving symlink destination: %w", err)
			}
		} else {
			absDest, err = filepath.Abs(dest)
			if err != nil {
				return fmt.Errorf("resolving symlink destination: %w", err)
			}
		}
	}

	if absDest != absSource {
		// Points somewhere else, leave it alone
		return nil
	}

	// Remove symlink
	if err := os.Remove(targetPath); err != nil {
		return fmt.Errorf("removing symlink: %w", err)
	}

	// Copy source file to target
	if err := l.copyFile(sourcePath, targetPath); err != nil {
		return fmt.Errorf("copying source file: %w", err)
	}

	return nil
}

func (l *Linker) copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
