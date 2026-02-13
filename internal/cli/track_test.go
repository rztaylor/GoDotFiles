package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestTrack(t *testing.T) {
	tmpDir := t.TempDir()
	home := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(home, ".gdf")

	// Setup environment
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatalf("creating home: %v", err)
	}
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", oldHome)

	// Configure git and init repo
	configureGitUserGlobal(t, home)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// Create a dummy dotfile
	dotfile := filepath.Join(home, ".testrc")
	content := "some config"
	if err := os.WriteFile(dotfile, []byte(content), 0644); err != nil {
		t.Fatalf("creating dotfile: %v", err)
	}

	// Track it (auto-detect app name 'testrc')
	if err := runTrack(nil, []string{dotfile}); err != nil {
		t.Fatalf("runTrack() error = %v", err)
	}

	// 1. Verify file moved to repo
	repoFile := filepath.Join(gdfDir, "dotfiles", "testrc", ".testrc")
	if _, err := os.Stat(repoFile); os.IsNotExist(err) {
		t.Error("file was not moved to repo")
	}

	// 2. Verify symlink created
	info, err := os.Lstat(dotfile)
	if err != nil {
		t.Fatalf("lstat dotfile: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("original file is not a symlink")
	}

	// 3. Verify app bundle created
	appPath := filepath.Join(gdfDir, "apps", "testrc.yaml")
	bundle, err := apps.Load(appPath)
	if err != nil {
		t.Fatalf("loading created bundle: %v", err)
	}
	if len(bundle.Dotfiles) != 1 {
		t.Errorf("dotfiles count = %d, want 1", len(bundle.Dotfiles))
	}
	if bundle.Dotfiles[0].Source != "testrc/.testrc" {
		t.Errorf("source = %q, want testrc/.testrc", bundle.Dotfiles[0].Source)
	}
	if bundle.Dotfiles[0].Target != "~/.testrc" {
		t.Errorf("target = %q, want ~/.testrc", bundle.Dotfiles[0].Target)
	}

	// Clean up for next test
	targetApp = ""
}

func TestTrackSecret(t *testing.T) {
	tmpDir := t.TempDir()
	home := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(home, ".gdf")

	os.Setenv("HOME", home)
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}

	configureGitUserGlobal(t, home)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatal(err)
	}

	// Create secret file
	secretFile := filepath.Join(home, ".aws", "credentials")
	if err := os.MkdirAll(filepath.Dir(secretFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secretFile, []byte("SECRET_KEY=123"), 0600); err != nil {
		t.Fatal(err)
	}

	// Track with --secret
	targetApp = "aws"
	secretFlag = true
	defer func() { secretFlag = false; targetApp = "" }()

	if err := runTrack(nil, []string{secretFile}); err != nil {
		t.Fatalf("runTrack(--secret) error = %v", err)
	}

	// 1. Verify file moved
	repoFile := filepath.Join(gdfDir, "dotfiles", "aws", "credentials")
	if _, err := os.Stat(repoFile); os.IsNotExist(err) {
		t.Error("file wasn't moved to repo")
	}

	// 2. Verify .gitignore updated
	gitignorePath := filepath.Join(gdfDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	expectedIgnore := "dotfiles/aws/credentials"
	if !strings.Contains(string(content), expectedIgnore) {
		t.Errorf(".gitignore missing entry %q. Content:\n%s", expectedIgnore, string(content))
	}

	// 3. Verify bundle has secret: true
	appPath := filepath.Join(gdfDir, "apps", "aws.yaml")
	bundle, err := apps.Load(appPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Dotfiles) != 1 {
		t.Fatal("dotfile not added to bundle")
	}
	if !bundle.Dotfiles[0].Secret {
		t.Error("dotfile.Secret should be true")
	}

	// 4. Verify idempotent (running again shouldn't duplicate in .gitignore)
	// We have to recreate the symlinked file first as if it was a new file,
	// or validly, runTrack fails if file already exists in repo.
	// Let's just check addToGitignore logic separately or trust the impl (which checks duplicates).
	// But let's verify duplicate check by calling addToGitignore directly
	if err := addToGitignore(gitignorePath, expectedIgnore); err != nil {
		t.Fatal(err)
	}
	content, _ = os.ReadFile(gitignorePath)
	count := strings.Count(string(content), expectedIgnore)
	if count != 1 {
		t.Errorf("expected 1 occurrence of %q in .gitignore, found %d", expectedIgnore, count)
	}
}

func TestAnalyzeApp(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/home/user/.gitconfig", "git"},
		{"/home/user/.zshrc", "zsh"},
		{"/home/user/.config/nvim/init.vim", "nvim"}, // dirname is nvim
		{"/home/user/unknown.conf", "unknown"},
	}

	for _, tt := range tests {
		got := apps.DetectAppFromPath(tt.path)
		if got != tt.want {
			t.Errorf("DetectAppFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
