package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository represents a Git repository.
type Repository struct {
	// Path is the root directory of the repository.
	Path string
}

// Init initializes a new Git repository at the given path.
func Init(path string) (*Repository, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("creating directory: %w", err)
	}

	// Run git init
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git init: %s - %w", string(output), err)
	}

	return &Repository{Path: path}, nil
}

// Clone clones a remote repository to the given path.
func Clone(url, path string) (*Repository, error) {
	cmd := exec.Command("git", "clone", url, path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git clone: %s - %w", string(output), err)
	}

	return &Repository{Path: path}, nil
}

// Open opens an existing Git repository.
func Open(path string) (*Repository, error) {
	// Check if .git exists
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not a git repository: %s", path)
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a git repository: %s", path)
	}

	return &Repository{Path: path}, nil
}

// IsRepository checks if a path is a Git repository.
func IsRepository(path string) bool {
	_, err := Open(path)
	return err == nil
}

// Add stages files for commit.
func (r *Repository) Add(paths ...string) error {
	args := append([]string{"add"}, paths...)
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %s - %w", string(output), err)
	}
	return nil
}

// Commit creates a commit with the given message.
func (r *Repository) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = r.Path
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check if nothing to commit
		if strings.Contains(string(output), "nothing to commit") {
			return nil
		}
		return fmt.Errorf("git commit: %s - %w", string(output), err)
	}
	return nil
}

// SetRemote sets the remote URL for the given remote name.
func (r *Repository) SetRemote(name, url string) error {
	// Try to add first
	cmd := exec.Command("git", "remote", "add", name, url)
	cmd.Dir = r.Path
	if _, err := cmd.CombinedOutput(); err != nil {
		// If it exists, set the URL
		cmd = exec.Command("git", "remote", "set-url", name, url)
		cmd.Dir = r.Path
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git remote set-url: %s - %w", string(output), err)
		}
	}
	return nil
}

// Push pushes commits to the remote repository.
func (r *Repository) Push() error {
	cmd := exec.Command("git", "push")
	cmd.Dir = r.Path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push: %s - %w", string(output), err)
	}
	return nil
}

// Pull pulls changes from the remote repository.
func (r *Repository) Pull() error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = r.Path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull: %s - %w", string(output), err)
	}
	return nil
}

// Status returns the current repository status.
func (r *Repository) Status() (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = r.Path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git status: %s - %w", string(output), err)
	}
	return string(output), nil
}

// HasChanges returns true if there are uncommitted changes.
func (r *Repository) HasChanges() (bool, error) {
	status, err := r.Status()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(status) != "", nil
}
