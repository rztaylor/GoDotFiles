package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/schema"
)

func TestApplyHooksDisabledByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	hookOutput := filepath.Join(homeDir, ".hook-disabled")
	b := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "hook-app",
		Hooks: &apps.Hooks{
			Apply: []apps.ApplyHook{{Run: "echo ran > " + hookOutput}},
		},
	}
	if err := b.Save(filepath.Join(gdfDir, "apps", "hook-app.yaml")); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	p, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	p.Apps = []string{"hook-app"}
	if err := p.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	oldRunHooks := applyRunHooks
	oldHookTimeout := applyHookTimeout
	defer func() {
		applyRunHooks = oldRunHooks
		applyHookTimeout = oldHookTimeout
	}()
	applyRunHooks = false
	applyHookTimeout = 2 * time.Second

	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply() error = %v", err)
	}
	if _, err := os.Stat(hookOutput); !os.IsNotExist(err) {
		t.Fatalf("expected hook output to be absent when hooks are not opted in, stat err=%v", err)
	}
}

func TestApplyRunsHooksWhenOptedIn(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	hookOutput := filepath.Join(homeDir, ".hook-enabled")
	b := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "hook-app",
		Hooks: &apps.Hooks{
			Apply: []apps.ApplyHook{{Run: "echo ran > " + hookOutput}},
		},
	}
	if err := b.Save(filepath.Join(gdfDir, "apps", "hook-app.yaml")); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	p, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	p.Apps = []string{"hook-app"}
	if err := p.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	oldRunHooks := applyRunHooks
	oldHookTimeout := applyHookTimeout
	defer func() {
		applyRunHooks = oldRunHooks
		applyHookTimeout = oldHookTimeout
	}()
	applyRunHooks = true
	applyHookTimeout = 2 * time.Second

	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply() error = %v", err)
	}
	if _, err := os.Stat(hookOutput); err != nil {
		t.Fatalf("expected hook output to exist, stat err=%v", err)
	}
}

func TestApplyHookTimeoutFails(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	b := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "slow-hook-app",
		Hooks: &apps.Hooks{
			Apply: []apps.ApplyHook{{Run: "sleep 1"}},
		},
	}
	if err := b.Save(filepath.Join(gdfDir, "apps", "slow-hook-app.yaml")); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	p, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	p.Apps = []string{"slow-hook-app"}
	if err := p.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	oldRunHooks := applyRunHooks
	oldHookTimeout := applyHookTimeout
	defer func() {
		applyRunHooks = oldRunHooks
		applyHookTimeout = oldHookTimeout
	}()
	applyRunHooks = true
	applyHookTimeout = 10 * time.Millisecond

	err = runApply(nil, []string{"default"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestApplyFailsWhenLockIsHeld(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	lockDir := filepath.Join(gdfDir, ".locks")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(lockDir, "apply.lock")
	if err := os.WriteFile(lockPath, []byte("held"), 0644); err != nil {
		t.Fatal(err)
	}

	err := runApply(nil, []string{"default"})
	if err == nil {
		t.Fatal("expected lock contention error")
	}
	if !strings.Contains(err.Error(), "already in progress") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyReleasesLockAfterSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(gdfDir, ".locks", "apply.lock")); !os.IsNotExist(err) {
		t.Fatalf("expected apply lock to be released, stat err=%v", err)
	}
}
