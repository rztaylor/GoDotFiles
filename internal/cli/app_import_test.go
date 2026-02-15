package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestDiscoverImportCandidates(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\nname = test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".aws"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".aws", "credentials"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".zshrc"), []byte("alias k='kubectl'\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dotfiles, aliases, err := discoverImportCandidates(home, nil)
	if err != nil {
		t.Fatalf("discoverImportCandidates() error = %v", err)
	}
	if len(dotfiles) == 0 {
		t.Fatal("expected at least one discovered dotfile")
	}
	if len(aliases) == 0 {
		t.Fatal("expected discovered aliases")
	}

	foundSensitive := false
	for _, d := range dotfiles {
		if d.Path == filepath.Join(home, ".aws", "credentials") {
			if !d.Sensitive {
				t.Fatal("aws credentials should be flagged sensitive")
			}
			foundSensitive = true
		}
	}
	if !foundSensitive {
		t.Fatal("expected to discover ~/.aws/credentials")
	}
}

func TestRunAppImport_ApplyRequiresSensitiveHandling(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configureGitUserGlobal(t, home)
	if err := createNewRepo(filepath.Join(home, ".gdf")); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(home, ".aws"), 0755); err != nil {
		t.Fatal(err)
	}
	sensitive := filepath.Join(home, ".aws", "credentials")
	if err := os.WriteFile(sensitive, []byte("aws_secret"), 0600); err != nil {
		t.Fatal(err)
	}

	old := importFlagSnapshot()
	defer restoreImportFlags(old)
	importApply = true
	importPreview = false
	importProfile = "default"
	importSensitiveHandling = ""

	err := runAppImport(nil, []string{sensitive})
	if err == nil {
		t.Fatal("expected error when applying sensitive import without handling")
	}
}

func TestRunAppImport_ApplySecretAndAlias(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configureGitUserGlobal(t, home)
	if err := createNewRepo(filepath.Join(home, ".gdf")); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(home, ".aws"), 0755); err != nil {
		t.Fatal(err)
	}
	sensitive := filepath.Join(home, ".aws", "credentials")
	if err := os.WriteFile(sensitive, []byte("aws_secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".zshrc"), []byte("alias gs='git status'\n"), 0644); err != nil {
		t.Fatal(err)
	}

	old := importFlagSnapshot()
	defer restoreImportFlags(old)
	importApply = true
	importPreview = false
	importProfile = "default"
	importSensitiveHandling = "secret"

	if err := runAppImport(nil, []string{sensitive}); err != nil {
		t.Fatalf("runAppImport() error = %v", err)
	}

	bundle, err := apps.Load(filepath.Join(home, ".gdf", "apps", "aws.yaml"))
	if err != nil {
		t.Fatalf("loading aws bundle: %v", err)
	}
	if len(bundle.Dotfiles) == 0 || !bundle.Dotfiles[0].Secret {
		t.Fatalf("expected secret dotfile in aws bundle, got %#v", bundle.Dotfiles)
	}

	aliasesPath := filepath.Join(home, ".gdf", "apps", "git.yaml")
	gitBundle, err := apps.Load(aliasesPath)
	if err != nil {
		t.Fatalf("loading git bundle for alias: %v", err)
	}
	if gitBundle.Shell == nil || gitBundle.Shell.Aliases["gs"] != "git status" {
		t.Fatalf("expected imported alias in git bundle, got %#v", gitBundle.Shell)
	}
}

type importFlags struct {
	preview           bool
	apply             bool
	json              bool
	profile           string
	sensitiveHandling string
}

func importFlagSnapshot() importFlags {
	return importFlags{
		preview:           importPreview,
		apply:             importApply,
		json:              importJSON,
		profile:           importProfile,
		sensitiveHandling: importSensitiveHandling,
	}
}

func restoreImportFlags(s importFlags) {
	importPreview = s.preview
	importApply = s.apply
	importJSON = s.json
	importProfile = s.profile
	importSensitiveHandling = s.sensitiveHandling
}
