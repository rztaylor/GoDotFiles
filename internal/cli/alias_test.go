package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
)

func TestAliasAdd(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	configureGitUserGlobal(t, homeDir)

	// Init repo
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatal(err)
	}

	// Create a kubectl app bundle so auto-detection can find it
	kubectlBundle := &apps.Bundle{
		Name:        "kubectl",
		Description: "Kubernetes CLI",
	}
	if err := kubectlBundle.Save(filepath.Join(gdfDir, "apps", "kubectl.yaml")); err != nil {
		t.Fatal(err)
	}

	// Test 1: Add alias with auto-detection (kubectl exists -> matches)
	args := []string{"k", "kubectl get pods"}
	aliasApp = "" // reset flag
	if err := runAliasAdd(nil, args); err != nil {
		t.Errorf("runAliasAdd(k) error = %v", err)
	}

	// Verify alias added to kubectl app
	appPath := filepath.Join(gdfDir, "apps", "kubectl.yaml")
	bundle, err := apps.Load(appPath)
	if err != nil {
		t.Fatal("kubectl app not found")
	}
	if bundle.Shell.Aliases["k"] != "kubectl get pods" {
		t.Errorf("alias k = %q, want 'kubectl get pods'", bundle.Shell.Aliases["k"])
	}

	// Test 2: Add alias with explicit app
	args = []string{"gco", "git checkout"}
	aliasApp = "git"
	if err := runAliasAdd(nil, args); err != nil {
		t.Errorf("runAliasAdd(gco) error = %v", err)
	}

	// Verify app 'git' created
	appPath = filepath.Join(gdfDir, "apps", "git.yaml")
	bundle, err = apps.Load(appPath)
	if err != nil {
		t.Fatal("git app not created")
	}
	if bundle.Shell.Aliases["gco"] != "git checkout" {
		t.Errorf("alias gco = %q, want 'git checkout'", bundle.Shell.Aliases["gco"])
	}
}

func TestAliasAddGlobal(t *testing.T) {
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

	// Add an alias for "ls -la" — no "ls" app exists, so it should go to global aliases
	aliasApp = ""
	if err := runAliasAdd(nil, []string{"ll", "ls -la"}); err != nil {
		t.Fatalf("runAliasAdd(ll) error = %v", err)
	}

	// Verify no "ls" app bundle was created
	lsAppPath := filepath.Join(gdfDir, "apps", "ls.yaml")
	if _, err := os.Stat(lsAppPath); !os.IsNotExist(err) {
		t.Error("ls app bundle should NOT be created for unmatched command")
	}

	// Verify alias went to global aliases file
	ga, err := apps.LoadGlobalAliases(filepath.Join(gdfDir, "aliases.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if ga.Aliases["ll"] != "ls -la" {
		t.Errorf("global alias ll = %q, want 'ls -la'", ga.Aliases["ll"])
	}
}

func TestAliasAddGlobal_PipelineCommand(t *testing.T) {
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

	// Pipeline command starting with cat — should go to global
	aliasApp = ""
	if err := runAliasAdd(nil, []string{"azsubs", "cat ~/.azure/subs.json | jq ."}); err != nil {
		t.Fatalf("runAliasAdd(azsubs) error = %v", err)
	}

	// No "cat" app should be created
	catAppPath := filepath.Join(gdfDir, "apps", "cat.yaml")
	if _, err := os.Stat(catAppPath); !os.IsNotExist(err) {
		t.Error("cat app bundle should NOT be created")
	}

	// Should be in global aliases
	ga, err := apps.LoadGlobalAliases(filepath.Join(gdfDir, "aliases.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if ga.Aliases["azsubs"] != "cat ~/.azure/subs.json | jq ." {
		t.Errorf("global alias azsubs = %q, want pipeline command", ga.Aliases["azsubs"])
	}
}

func TestAliasConflict(t *testing.T) {
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

	// Add alias first time
	aliasApp = "git"
	if err := runAliasAdd(nil, []string{"gco", "git checkout"}); err != nil {
		t.Fatal(err)
	}

	// Add same alias with different command (should overwrite)
	if err := runAliasAdd(nil, []string{"gco", "git commit"}); err != nil {
		t.Fatal(err)
	}

	// Verify overwrite happened
	appPath := filepath.Join(gdfDir, "apps", "git.yaml")
	bundle, err := apps.Load(appPath)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Shell.Aliases["gco"] != "git commit" {
		t.Errorf("alias gco = %q, want 'git commit' (overwrite)", bundle.Shell.Aliases["gco"])
	}
}

func TestAliasRemove(t *testing.T) {
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

	// Create app with alias
	bundle := &apps.Bundle{
		Name: "git",
		Shell: &apps.Shell{
			Aliases: map[string]string{
				"gco": "git checkout",
			},
		},
	}
	appPath := filepath.Join(gdfDir, "apps", "git.yaml")
	if err := bundle.Save(appPath); err != nil {
		t.Fatal(err)
	}

	// Remove alias — no profile needed, searches all apps
	args := []string{"gco"}
	if err := runAliasRemove(nil, args); err != nil {
		t.Errorf("runAliasRemove(gco) error = %v", err)
	}

	// Verify removed
	bundle, err := apps.Load(appPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := bundle.Shell.Aliases["gco"]; ok {
		t.Error("alias gco was not removed")
	}
}

func TestAliasRemoveGlobal(t *testing.T) {
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

	// Add a global alias
	ga := &apps.GlobalAliases{
		Aliases: map[string]string{
			"ll": "ls -la",
			"..": "cd ..",
		},
	}
	if err := ga.Save(filepath.Join(gdfDir, "aliases.yaml")); err != nil {
		t.Fatal(err)
	}

	// Remove global alias
	if err := runAliasRemove(nil, []string{"ll"}); err != nil {
		t.Errorf("runAliasRemove(ll) should succeed: %v", err)
	}

	// Verify removed
	loaded, err := apps.LoadGlobalAliases(filepath.Join(gdfDir, "aliases.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := loaded.Aliases["ll"]; ok {
		t.Error("global alias 'll' was not removed")
	}
	// Other alias still intact
	if loaded.Aliases[".."] != "cd .." {
		t.Error("global alias '..' should still exist")
	}
}

func TestAliasRemoveMultiApp(t *testing.T) {
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

	// Create multiple apps with the same alias name
	for _, app := range []string{"git", "kubectl", "docker"} {
		bundle := &apps.Bundle{
			Name: app,
			Shell: &apps.Shell{
				Aliases: map[string]string{
					"test": app + " test",
				},
			},
		}
		appPath := filepath.Join(gdfDir, "apps", app+".yaml")
		if err := bundle.Save(appPath); err != nil {
			t.Fatal(err)
		}
	}

	// Remove alias (should remove from first match found)
	if err := runAliasRemove(nil, []string{"test"}); err != nil {
		t.Fatal(err)
	}

	// Verify removed from one app (the first alphabetically since we use ReadDir)
	dockerBundle, _ := apps.Load(filepath.Join(gdfDir, "apps", "docker.yaml"))
	if _, ok := dockerBundle.Shell.Aliases["test"]; ok {
		t.Error("alias test was not removed from docker (first alpha match)")
	}

	// Still exists in others
	kubectlBundle, _ := apps.Load(filepath.Join(gdfDir, "apps", "kubectl.yaml"))
	if _, ok := kubectlBundle.Shell.Aliases["test"]; !ok {
		t.Error("alias test should still exist in kubectl")
	}
}

func TestAliasRemoveNotFound(t *testing.T) {
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

	// Try to remove non-existent alias
	err := runAliasRemove(nil, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error when removing non-existent alias")
	}
}

func TestAliasList(t *testing.T) {
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

	// Empty state — should not error
	if err := runAliasList(nil, nil); err != nil {
		t.Errorf("runAliasList error = %v", err)
	}
}

func TestAliasListShowsAllAppsAndGlobal(t *testing.T) {
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

	// Create apps with aliases — NOT necessarily in any profile
	appsData := map[string]map[string]string{
		"git":     {"gco": "git checkout", "gst": "git status"},
		"kubectl": {"k": "kubectl", "kgp": "kubectl get pods"},
	}

	for appName, aliases := range appsData {
		bundle := &apps.Bundle{
			Name: appName,
			Shell: &apps.Shell{
				Aliases: aliases,
			},
		}
		appPath := filepath.Join(gdfDir, "apps", appName+".yaml")
		if err := bundle.Save(appPath); err != nil {
			t.Fatal(err)
		}
	}

	// Add global aliases
	ga := &apps.GlobalAliases{
		Aliases: map[string]string{
			"ll": "ls -la",
			"..": "cd ..",
		},
	}
	if err := ga.Save(filepath.Join(gdfDir, "aliases.yaml")); err != nil {
		t.Fatal(err)
	}

	// List should work without error and show everything
	if err := runAliasList(nil, nil); err != nil {
		t.Fatal(err)
	}
}

// TestAliasAddAndRemoveRoundTrip tests the full workflow: add an alias for an
// unknown command, then remove it. The old code would fail because removing
// only searched within a profile's apps.
func TestAliasAddAndRemoveRoundTrip(t *testing.T) {
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

	// Create a profile with git in it
	profile := &config.Profile{Name: "base", Apps: []string{"git"}}
	profileDir := filepath.Join(gdfDir, "profiles", "base")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := profile.Save(filepath.Join(profileDir, "profile.yaml")); err != nil {
		t.Fatal(err)
	}

	gitBundle := &apps.Bundle{
		Name:  "git",
		Shell: &apps.Shell{Aliases: map[string]string{"g": "git"}},
	}
	if err := gitBundle.Save(filepath.Join(gdfDir, "apps", "git.yaml")); err != nil {
		t.Fatal(err)
	}

	// Add alias for "ls -la" — should go to global aliases
	aliasApp = ""
	if err := runAliasAdd(nil, []string{"ll", "ls -la"}); err != nil {
		t.Fatalf("add error: %v", err)
	}

	// Verify no ls.yaml created
	if _, err := os.Stat(filepath.Join(gdfDir, "apps", "ls.yaml")); !os.IsNotExist(err) {
		t.Fatal("ls.yaml should not exist")
	}

	// Remove alias — should find it in global aliases
	if err := runAliasRemove(nil, []string{"ll"}); err != nil {
		t.Fatalf("remove error: %v", err)
	}

	// Verify removed
	ga, err := apps.LoadGlobalAliases(filepath.Join(gdfDir, "aliases.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ga.Aliases["ll"]; ok {
		t.Error("alias 'll' should have been removed from global aliases")
	}
}
