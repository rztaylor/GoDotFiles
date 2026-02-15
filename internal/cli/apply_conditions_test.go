package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/schema"
)

func TestApplyDotfileConditions(t *testing.T) {
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

	// Force platform for deterministic condition results.
	previous := platform.Override
	platform.Override = &platform.Platform{OS: "linux", Distro: "ubuntu", Hostname: "test-host", Arch: "amd64", Home: homeDir}
	defer func() {
		platform.Override = previous
	}()

	// Source files
	linuxSource := filepath.Join(gdfDir, "dotfiles", "testapp", "linux")
	macosSource := filepath.Join(gdfDir, "dotfiles", "testapp", "macos")
	for _, p := range []string{linuxSource, macosSource} {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	bundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "testapp",
		Dotfiles: []apps.Dotfile{
			{
				Source: "testapp/linux",
				Target: "~/.config/testapp/active",
				When:   "os == linux OR os == wsl",
			},
			{
				Source: "testapp/macos",
				Target: "~/.config/testapp/active",
				When:   "os == macos",
			},
		},
	}
	if err := bundle.Save(filepath.Join(gdfDir, "apps", "testapp.yaml")); err != nil {
		t.Fatal(err)
	}

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = []string{"testapp"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply: %v", err)
	}

	target := filepath.Join(homeDir, ".config", "testapp", "active")
	info, err := os.Lstat(target)
	if err != nil {
		t.Fatalf("lstat target: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("target is not symlink")
	}
	dest, _ := os.Readlink(target)
	if dest != linuxSource {
		t.Errorf("symlink destination = %q, want %q", dest, linuxSource)
	}
}

func TestApplyDotfileConditionInvalid(t *testing.T) {
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

	sourcePath := filepath.Join(gdfDir, "dotfiles", "testapp", "cfg")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	bundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "testapp",
		Dotfiles: []apps.Dotfile{
			{
				Source: "testapp/cfg",
				Target: "~/.testapp",
				When:   "os ==",
			},
		},
	}
	if err := bundle.Save(filepath.Join(gdfDir, "apps", "testapp.yaml")); err != nil {
		t.Fatal(err)
	}

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = []string{"testapp"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	if err := runApply(nil, []string{"default"}); err == nil {
		t.Fatal("expected runApply to fail for invalid dotfile condition")
	}
}

func TestApplyDotfileTargetMap(t *testing.T) {
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

	previous := platform.Override
	platform.Override = &platform.Platform{OS: "linux", Distro: "ubuntu", Hostname: "test-host", Arch: "amd64", Home: homeDir}
	defer func() {
		platform.Override = previous
	}()

	sourcePath := filepath.Join(gdfDir, "dotfiles", "testapp", "cfg")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Build via YAML to exercise target-map decode.
	appYAML := `
kind: App/v1
name: testapp
dotfiles:
  - source: testapp/cfg
    target:
      default: ~/.config/testapp/default
      linux: ~/.config/testapp/linux
`
	if err := os.WriteFile(filepath.Join(gdfDir, "apps", "testapp.yaml"), []byte(appYAML), 0644); err != nil {
		t.Fatal(err)
	}

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = []string{"testapp"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply: %v", err)
	}

	target := filepath.Join(homeDir, ".config", "testapp", "linux")
	info, err := os.Lstat(target)
	if err != nil {
		t.Fatalf("lstat target: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("target is not symlink")
	}
}
