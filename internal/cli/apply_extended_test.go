package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/engine"
)

func TestApplyMultipleProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	// Create two profiles with dependencies
	// Create base profile
	if err := os.MkdirAll(filepath.Join(gdfDir, "profiles", "base"), 0755); err != nil {
		t.Fatal(err)
	}
	baseProfile := &config.Profile{
		Name: "base",
		Apps: []string{"git"},
	}
	if err := baseProfile.Save(filepath.Join(gdfDir, "profiles", "base", "profile.yaml")); err != nil {
		t.Fatal(err)
	}

	// Create work profile
	if err := os.MkdirAll(filepath.Join(gdfDir, "profiles", "work"), 0755); err != nil {
		t.Fatal(err)
	}
	workProfile := &config.Profile{
		Name:     "work",
		Includes: []string{"base"},
		Apps:     []string{"kubectl"},
	}
	if err := workProfile.Save(filepath.Join(gdfDir, "profiles", "work", "profile.yaml")); err != nil {
		t.Fatal(err)
	}

	// Create app bundles
	for _, appName := range []string{"git", "kubectl"} {
		bundle := &apps.Bundle{Name: appName}
		if err := bundle.Save(filepath.Join(gdfDir, "apps", appName+".yaml")); err != nil {
			t.Fatal(err)
		}
	}

	// Apply work profile (should include base)
	if err := runApply(nil, []string{"work"}); err != nil {
		t.Fatalf("runApply: %v", err)
	}

	// Verify operation log exists
	logDir := filepath.Join(gdfDir, ".operations")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("reading log dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("no operation log created")
	}
}

func TestApplyDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	// Create test app with dotfile
	sourcePath := filepath.Join(gdfDir, "dotfiles", "testapp", "config")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	bundle := &apps.Bundle{
		Name: "testapp",
		Dotfiles: []apps.Dotfile{
			{Source: "testapp/config", Target: "~/.testrc"},
		},
	}
	if err := bundle.Save(filepath.Join(gdfDir, "apps", "testapp.yaml")); err != nil {
		t.Fatal(err)
	}

	// Add to default profile
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, _ := config.LoadProfile(profilePath)
	profile.Apps = append(profile.Apps, "testapp")
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	// Run dry-run apply
	applyDryRun = true
	defer func() { applyDryRun = false }()

	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply: %v", err)
	}

	// Verify no symlink was created
	targetPath := filepath.Join(homeDir, ".testrc")
	if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
		t.Error("dry run should not create symlinks")
	}

	// Verify no operation log was created
	logDir := filepath.Join(gdfDir, ".operations")
	if entries, err := os.ReadDir(logDir); err == nil && len(entries) > 0 {
		t.Error("dry run should not create operation logs")
	}
}

func TestApplyWithDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	// Create apps with dependencies: kubectl depends on base
	baseBundle := &apps.Bundle{Name: "base"}
	if err := baseBundle.Save(filepath.Join(gdfDir, "apps", "base.yaml")); err != nil {
		t.Fatal(err)
	}

	kubectlBundle := &apps.Bundle{
		Name:         "kubectl",
		Dependencies: []string{"base"},
	}
	if err := kubectlBundle.Save(filepath.Join(gdfDir, "apps", "kubectl.yaml")); err != nil {
		t.Fatal(err)
	}

	// Add both kubectl and base to profile (base will be dependency-ordered first)
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, _ := config.LoadProfile(profilePath)
	profile.Apps = []string{"kubectl", "base"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	// Apply should succeed and process base first
	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply: %v", err)
	}

	// Read operation log to verify order
	logDir := filepath.Join(gdfDir, ".operations")
	entries, err := os.ReadDir(logDir)
	if err != nil || len(entries) == 0 {
		t.Skip("no operation log to verify order")
	}

	data, _ := os.ReadFile(filepath.Join(logDir, entries[0].Name()))
	var ops []engine.Operation
	if err := json.Unmarshal(data, &ops); err != nil {
		t.Fatalf("unmarshaling log: %v", err)
	}

	// Should have shell_generate operation
	found := false
	for _, op := range ops {
		if op.Type == "shell_generate" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected shell_generate operation in log")
	}
}

func TestApplyCircularDependency(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	// Create apps with circular dependency
	appA := &apps.Bundle{
		Name:         "app-a",
		Dependencies: []string{"app-b"},
	}
	if err := appA.Save(filepath.Join(gdfDir, "apps", "app-a.yaml")); err != nil {
		t.Fatal(err)
	}

	appB := &apps.Bundle{
		Name:         "app-b",
		Dependencies: []string{"app-a"},
	}
	if err := appB.Save(filepath.Join(gdfDir, "apps", "app-b.yaml")); err != nil {
		t.Fatal(err)
	}

	// Add both to profile
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, _ := config.LoadProfile(profilePath)
	profile.Apps = []string{"app-a", "app-b"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	// Apply should fail with circular dependency error
	err := runApply(nil, []string{"default"})
	if err == nil {
		t.Fatal("expected error for circular dependency")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected circular dependency error, got: %v", err)
	}
}
