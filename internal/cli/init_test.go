package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/git"
)

func TestCreateNewRepo(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")

	// Set HOME to tmpDir for the test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Configure git user globally for test
	configureGitUserGlobal(t, tmpDir)

	err := createNewRepo(gdfDir)
	if err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// Check that it's a git repo
	if !git.IsRepository(gdfDir) {
		t.Error("directory is not a git repository")
	}

	// Check directories exist
	dirs := []string{"apps", "profiles", "dotfiles"}
	for _, dir := range dirs {
		path := filepath.Join(gdfDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("directory %s was not created", dir)
		}
	}

	// Check files exist
	files := []string{".gitignore", "config.yaml"}
	for _, file := range files {
		path := filepath.Join(gdfDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file %s was not created", file)
		}
	}

	// Check default profile exists
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		t.Error("default profile was not created")
	}

	// Check generated init placeholder exists
	generatedInitPath := filepath.Join(gdfDir, "generated", "init.sh")
	content, err := os.ReadFile(generatedInitPath)
	if err != nil {
		t.Fatalf("reading generated init placeholder: %v", err)
	}
	if !containsString(string(content), "Placeholder created by gdf init.") {
		t.Errorf("generated init placeholder content missing marker:\n%s", string(content))
	}

	// Check config includes full default sections.
	configPath := filepath.Join(gdfDir, "config.yaml")
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config.yaml: %v", err)
	}
	requiredConfigFields := []string{
		"git:",
		"package_manager:",
		"updates:",
		"shell_integration:",
		"check_interval: 24h",
	}
	for _, field := range requiredConfigFields {
		if !containsString(string(configContent), field) {
			t.Errorf("config.yaml missing field: %s", field)
		}
	}
}

func TestCreateGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	err := createGitignore(tmpDir)
	if err != nil {
		t.Fatalf("createGitignore() error = %v", err)
	}

	// Check file exists
	path := filepath.Join(tmpDir, ".gitignore")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading .gitignore: %v", err)
	}

	// Check required entries
	required := []string{"state.yaml", ".operations/", ".history/"}
	for _, entry := range required {
		if !containsString(string(content), entry) {
			t.Errorf(".gitignore missing entry: %s", entry)
		}
	}
}

func TestCreateDefaultProfile(t *testing.T) {
	tmpDir := t.TempDir()

	err := createDefaultProfile(tmpDir)
	if err != nil {
		t.Fatalf("createDefaultProfile() error = %v", err)
	}

	// Check profile directory exists
	profileDir := filepath.Join(tmpDir, "profiles", "default")
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		t.Error("default profile directory was not created")
	}

	// Check profile.yaml exists
	profilePath := filepath.Join(profileDir, "profile.yaml")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("reading profile.yaml: %v", err)
	}

	if !containsString(string(content), "name: default") {
		t.Error("profile.yaml missing name field")
	}
}

func TestEnsureGeneratedInitScript_Idempotent(t *testing.T) {
	gdfDir := t.TempDir()
	scriptPath := filepath.Join(gdfDir, "generated", "init.sh")

	customContent := "# custom\n"
	if err := os.MkdirAll(filepath.Dir(scriptPath), 0755); err != nil {
		t.Fatalf("creating generated dir: %v", err)
	}
	if err := os.WriteFile(scriptPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("writing existing script: %v", err)
	}

	if err := ensureGeneratedInitScript(gdfDir); err != nil {
		t.Fatalf("ensureGeneratedInitScript() error = %v", err)
	}

	got, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("reading script after ensure: %v", err)
	}
	if string(got) != customContent {
		t.Fatalf("script content changed; got %q, want %q", string(got), customContent)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// configureGitUserGlobal sets up git user config in a test-specific directory
func configureGitUserGlobal(t *testing.T, home string) {
	t.Helper()

	gitconfigPath := filepath.Join(home, ".gitconfig")
	content := `[user]
	email = test@example.com
	name = Test User
`
	if err := os.WriteFile(gitconfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("writing .gitconfig: %v", err)
	}
}
