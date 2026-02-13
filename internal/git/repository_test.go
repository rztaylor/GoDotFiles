package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo, err := Init(repoPath)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if repo.Path != repoPath {
		t.Errorf("Path = %q, want %q", repo.Path, repoPath)
	}

	// Check .git directory exists
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory was not created")
	}
}

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a repo first
	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Open existing repo
	opened, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if opened.Path != repo.Path {
		t.Errorf("Path = %q, want %q", opened.Path, repo.Path)
	}

	// Try to open non-repo
	_, err = Open(t.TempDir())
	if err == nil {
		t.Error("Open() expected error for non-repo, got nil")
	}
}

func TestIsRepository(t *testing.T) {
	tmpDir := t.TempDir()

	// Before init
	if IsRepository(tmpDir) {
		t.Error("IsRepository() = true before init, want false")
	}

	// After init
	if _, err := Init(tmpDir); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if !IsRepository(tmpDir) {
		t.Error("IsRepository() = false after init, want true")
	}
}

func TestRepository_AddCommit(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Configure git user for commit
	configureGitUser(t, tmpDir)

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	// Add file
	if err := repo.Add("test.txt"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Commit
	if err := repo.Commit("test commit"); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Should have no changes after commit
	hasChanges, err := repo.HasChanges()
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if hasChanges {
		t.Error("HasChanges() = true after commit, want false")
	}
}

func TestRepository_HasChanges(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Initially no changes
	hasChanges, err := repo.HasChanges()
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if hasChanges {
		t.Error("HasChanges() = true for empty repo, want false")
	}

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	// Should detect untracked file
	hasChanges, err = repo.HasChanges()
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if !hasChanges {
		t.Error("HasChanges() = false with untracked file, want true")
	}
}

// configureGitUser sets up git user for tests
func configureGitUser(t *testing.T, dir string) {
	t.Helper()

	// Set local git config for test repo
	cmds := [][]string{
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
	}

	for _, cmd := range cmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Fatalf("configuring git: %v", err)
		}
	}
}
