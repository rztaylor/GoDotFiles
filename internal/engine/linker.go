package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

// Linker handles the creation and management of symlinks for dotfiles.
type Linker struct {
	// ConflictStrategy determines how to handle existing files at the target location.
	// Options: "backup_and_replace", "replace", "error", "prompt" (prompt not implemented here, assumed resolved upstream)
	ConflictStrategy string

	history           *HistoryManager
	conflictSnapshots map[string]*Snapshot
}

// NewLinker creates a new Linker with the given conflict strategy.
func NewLinker(strategy string) *Linker {
	return &Linker{
		ConflictStrategy:  strategy,
		conflictSnapshots: make(map[string]*Snapshot),
	}
}

// SetHistoryManager configures snapshot capture for destructive operations.
func (l *Linker) SetHistoryManager(history *HistoryManager) {
	l.history = history
}

// ConsumeConflictSnapshot returns and clears the snapshot captured for target.
func (l *Linker) ConsumeConflictSnapshot(target string) *Snapshot {
	s := l.conflictSnapshots[target]
	delete(l.conflictSnapshots, target)
	return s
}

// Link processes a single dotfile, creating a symlink from target to source.
// source: path relative to repo root (e.g. "git/.gitconfig")
// target: absolute path or path relative to home (e.g. "~/.gitconfig")
// gdfDir: absolute path to GDF repo root
func (l *Linker) Link(dotfile apps.Dotfile, gdfDir string) error {
	sourcePath := filepath.Join(gdfDir, "dotfiles", dotfile.Source)
	targetPath := platform.ExpandPath(dotfile.Target)

	// Handle secrets (skip or warn?)
	if dotfile.Secret {
		fmt.Printf("Warning: Linking secret file %s - ensure it's gitignored\n", dotfile.Target)
	}

	// Validate source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", sourcePath)
	}

	// Check if directory structure exists for target
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("creating directory for target: %w", err)
	}

	// Check target state
	info, err := os.Lstat(targetPath)
	if err == nil {
		// Target exists
		// 1. Check if it's already a symlink to the correct source
		if info.Mode()&os.ModeSymlink != 0 {
			linkDest, err := os.Readlink(targetPath)
			if err != nil {
				return fmt.Errorf("reading symlink: %w", err)
			}
			if linkDest == sourcePath {
				// Already linked correctly
				return nil
			}
		}

		// 2. Handle conflict
		snapshot, err := l.handleConflict(targetPath)
		if err != nil {
			return err
		}
		if snapshot != nil {
			l.conflictSnapshots[targetPath] = snapshot
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking target: %w", err)
	}

	// Create symlink
	if err := os.Symlink(sourcePath, targetPath); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}

	return nil
}

// Unlink removes the symlink for a dotfile.
func (l *Linker) Unlink(dotfile apps.Dotfile) error {
	_, err := l.UnlinkManaged(dotfile, "")
	return err
}

// UnlinkManaged removes a managed symlink and returns a snapshot for rollback.
// If gdfDir is provided, only symlinks pointing at the managed dotfile source are removed.
func (l *Linker) UnlinkManaged(dotfile apps.Dotfile, gdfDir string) (*Snapshot, error) {
	targetPath := platform.ExpandPath(dotfile.Target)

	info, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Already gone
		}
		return nil, err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		// If it's not a symlink, we leave it alone (it might be a real file)
		return nil, nil
	}

	if gdfDir != "" && dotfile.Source != "" {
		expectedSource := filepath.Join(gdfDir, "dotfiles", dotfile.Source)
		actualSource, err := resolveSymlinkDestination(targetPath)
		if err != nil {
			return nil, fmt.Errorf("reading symlink destination: %w", err)
		}
		if actualSource != expectedSource {
			return nil, nil
		}
	}

	snapshot, err := l.captureSnapshot(targetPath)
	if err != nil {
		return nil, err
	}
	if err := os.Remove(targetPath); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func resolveSymlinkDestination(targetPath string) (string, error) {
	dest, err := os.Readlink(targetPath)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(dest) {
		return filepath.Clean(dest), nil
	}
	return filepath.Clean(filepath.Join(filepath.Dir(targetPath), dest)), nil
}

func (l *Linker) handleConflict(path string) (*Snapshot, error) {
	var snapshot *Snapshot

	switch l.ConflictStrategy {
	case "error":
		return nil, fmt.Errorf("target already exists: %s", path)
	case "replace", "force":
		s, err := l.captureSnapshot(path)
		if err != nil {
			return nil, err
		}
		snapshot = s
		if err := os.RemoveAll(path); err != nil {
			return nil, fmt.Errorf("removing existing target: %w", err)
		}
	case "backup_and_replace":
		s, err := l.captureSnapshot(path)
		if err != nil {
			return nil, err
		}
		snapshot = s

		// Cycle backups: .bak -> .bak.1 -> .bak.2 -> .bak.3 (keep last 3)
		const maxBackups = 3
		for i := maxBackups - 1; i >= 0; i-- {
			var oldPath, newPath string
			if i == 0 {
				oldPath = path + ".gdf.bak"
			} else {
				oldPath = fmt.Sprintf("%s.gdf.bak.%d", path, i)
			}
			newPath = fmt.Sprintf("%s.gdf.bak.%d", path, i+1)

			if _, err := os.Stat(oldPath); err == nil {
				if i == maxBackups-1 {
					// Remove oldest backup
					_ = os.Remove(oldPath)
				} else {
					// Rename to next number
					_ = os.Rename(oldPath, newPath)
				}
			}
		}
		// Move current file to .bak
		backupPath := path + ".gdf.bak"
		if err := os.Rename(path, backupPath); err != nil {
			return nil, fmt.Errorf("backing up existing target: %w", err)
		}
	default:
		// Default to error for safety
		return nil, fmt.Errorf("unknown conflict strategy '%s', defaulting to error: file exists %s", l.ConflictStrategy, path)
	}
	return snapshot, nil
}

func (l *Linker) captureSnapshot(path string) (*Snapshot, error) {
	if l.history == nil {
		return nil, nil
	}
	s, err := l.history.Capture(path)
	if err != nil {
		return nil, fmt.Errorf("capturing history snapshot for %s: %w", path, err)
	}
	return s, nil
}
