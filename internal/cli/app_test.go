package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddApp(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")

	// Mock environment
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Configure git user
	configureGitUserGlobal(t, tmpDir)

	// Initialize repo
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// Helper to run add command
	runAddCmd := func(app, profile string) error {
		targetProfile = profile
		return runAdd(nil, []string{app})
	}

	// 1. Add new app to default profile
	if err := runAddCmd("kubectl", "default"); err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}

	// Verify app file created
	appPath := filepath.Join(gdfDir, "apps", "kubectl.yaml")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		t.Error("app file was not created")
	}

	// Verify added to profile
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("reading profile: %v", err)
	}
	if !containsString(string(content), "- kubectl") {
		t.Error("app not added to profile")
	}

	// 2. Add existing app to new profile
	if err := runAddCmd("kubectl", "work"); err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}

	// Verify created new profile
	workProfilePath := filepath.Join(gdfDir, "profiles", "work", "profile.yaml")
	if _, err := os.Stat(workProfilePath); os.IsNotExist(err) {
		t.Error("new profile was not created")
	}

	// 3. Add duplicate app (should be idempotent)
	if err := runAddCmd("kubectl", "default"); err != nil {
		t.Fatalf("runAdd() duplicate error = %v", err)
	}
}

func TestRemoveApp(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	os.Setenv("HOME", tmpDir)

	configureGitUserGlobal(t, tmpDir)

	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// Add app first
	targetProfile = "default"
	if err := runAdd(nil, []string{"git"}); err != nil {
		t.Fatalf("setup: runAdd() error = %v", err)
	}

	// Remove app
	if err := runRemove(nil, []string{"git"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}

	// Verify removed from profile
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("reading profile: %v", err)
	}
	if containsString(string(content), "- git") {
		t.Error("app was not removed from profile")
	}

	// Remove non-existent app (should perform no-op)
	if err := runRemove(nil, []string{"missing"}); err != nil {
		t.Fatalf("runRemove() missing error = %v", err)
	}
}

func TestListApps(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	os.Setenv("HOME", tmpDir)

	configureGitUserGlobal(t, tmpDir)

	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// Add apps
	targetProfile = "default"
	_ = runAdd(nil, []string{"app1"})
	_ = runAdd(nil, []string{"app2"})

	// Verify apps were added to profile
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("reading profile: %v", err)
	}

	profileStr := string(content)
	if !containsString(profileStr, "- app1") {
		t.Error("app1 not found in profile output")
	}
	if !containsString(profileStr, "- app2") {
		t.Error("app2 not found in profile output")
	}

	// Run list command
	if err := runList(nil, nil); err != nil {
		t.Fatalf("runList() error = %v", err)
	}
}
