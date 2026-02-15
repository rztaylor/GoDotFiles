package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/stretchr/testify/assert"
)

func TestInstall_Learning(t *testing.T) {
	// Setup repo
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatal(err)
	}

	// Case 1: Install new app, learn package
	// We need to simulate stdin/stdout interaction.
	// Since runInstall reads from os.Stdin, we can pipe input.

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	// Input sequence:
	// 1. Install Prompt: "1" (Select Package Manager)
	// 2. Package Name: "my-pkg-name"
	go func() {
		defer w.Close()
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte("1\n")) // Select PM
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte("my-pkg-name\n")) // Pkg Name
	}()

	// Reset installProfile global.
	installProfile = ""
	defer func() { installProfile = "" }()

	// Override platform to ensure deterministic behavior (simulate Linux)
	platform.Override = &platform.Platform{
		OS:     "linux",
		Distro: "ubuntu",
		Arch:   "amd64",
		Home:   homeDir,
	}
	defer func() { platform.Override = nil }()

	// Override package manager to prevent sudo
	packages.Override = &MockPackageManager{mgrName: "apt"}
	defer func() { packages.Override = nil }()

	err = runInstall(nil, []string{"new-app"})
	assert.NoError(t, err)

	// Check if app bundle was created
	appPath := filepath.Join(gdfDir, "apps", "new-app.yaml")
	assert.FileExists(t, appPath)

	bundle, err := apps.Load(appPath)
	if err != nil {
		t.Fatalf("failed to load bundle: %v", err)
	}
	assert.Equal(t, "new-app", bundle.Name)

	// Verify package was learned
	// We forced Linux/Ubuntu, so we expect Apt configuration.
	if assert.NotNil(t, bundle.Package) {
		if assert.NotNil(t, bundle.Package.Apt) {
			assert.Equal(t, "my-pkg-name", bundle.Package.Apt.Name)
		}
	}

	// Verify profile was created and app added.
	// The repo has a single profile by default, so omitted --profile resolves to that profile.
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	assert.FileExists(t, profilePath)

	p, err := config.LoadProfile(profilePath)
	assert.NoError(t, err)
	assert.Contains(t, p.Apps, "new-app")
}

func TestInstall_WithProfileFlag(t *testing.T) {
	// Setup repo
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	createNewRepo(gdfDir)

	// Mock Stdin for just the package learning part (profile should be skipped)
	r, w, _ := os.Pipe()
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	go func() {
		defer w.Close()
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte("1\n")) // Select PM
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte("pkg-for-cli\n")) // Pkg Name
	}()

	platform.Override = &platform.Platform{OS: "linux", Distro: "ubuntu", Home: homeDir}
	defer func() { platform.Override = nil }()
	packages.Override = &MockPackageManager{mgrName: "apt"}
	defer func() { packages.Override = nil }()

	// Set the global flag variable for the test
	installProfile = "cli-profile"
	defer func() { installProfile = "" }()

	err := runInstall(nil, []string{"cli-app"})
	assert.NoError(t, err)

	// Verify profile created
	profilePath := filepath.Join(gdfDir, "profiles", "cli-profile", "profile.yaml")
	assert.FileExists(t, profilePath)
}

// MockPackageManager for testing
type MockPackageManager struct {
	mgrName     string
	installed   []string
	uninstalled []string
}

func (m *MockPackageManager) Name() string {
	return m.mgrName
}

func (m *MockPackageManager) Install(pkg string) error {
	m.installed = append(m.installed, pkg)
	return nil
}

func (m *MockPackageManager) Uninstall(pkg string) error {
	m.uninstalled = append(m.uninstalled, pkg)
	return nil
}

func (m *MockPackageManager) IsInstalled(pkg string) (bool, error) {
	for _, p := range m.installed {
		if p == pkg {
			return true, nil
		}
	}
	return false, nil
}
