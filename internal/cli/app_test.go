package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/schema"
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

func TestAddAppCreatesAppsDirectoryWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	if err := os.RemoveAll(filepath.Join(gdfDir, "apps")); err != nil {
		t.Fatalf("removing apps dir: %v", err)
	}

	oldProfile := targetProfile
	oldFromRecipe := fromRecipe
	targetProfile = "base"
	fromRecipe = true
	defer func() {
		targetProfile = oldProfile
		fromRecipe = oldFromRecipe
	}()

	if err := runAdd(nil, []string{"git"}); err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(gdfDir, "apps", "git.yaml")); err != nil {
		t.Fatalf("expected git app bundle to be created: %v", err)
	}
}

func TestAddAppFromRecipeSeedsDotfileSources(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	oldProfile := targetProfile
	oldFromRecipe := fromRecipe
	targetProfile = "default"
	fromRecipe = true
	defer func() {
		targetProfile = oldProfile
		fromRecipe = oldFromRecipe
	}()

	tests := []struct {
		appName string
		source  string
		marker  string
	}{
		{appName: "git", source: "gitconfig", marker: "Managed by gdf recipe 'git'"},
		{appName: "zsh", source: "zshrc", marker: "Managed by gdf recipe 'zsh'"},
		{appName: "oh-my-zsh", source: "oh-my-zsh/custom/aliases.zsh", marker: "Managed by gdf recipe 'oh-my-zsh'"},
	}

	for _, tt := range tests {
		if err := runAdd(nil, []string{tt.appName}); err != nil {
			t.Fatalf("runAdd(%q) error = %v", tt.appName, err)
		}
		content, err := os.ReadFile(filepath.Join(gdfDir, "dotfiles", tt.source))
		if err != nil {
			t.Fatalf("expected seeded source for %q: %v", tt.appName, err)
		}
		if !strings.Contains(string(content), tt.marker) {
			t.Fatalf("seeded source for %q missing marker %q", tt.appName, tt.marker)
		}
	}
}

func TestAddAppFromRecipeKeepsExistingDotfileSource(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	existingPath := filepath.Join(gdfDir, "dotfiles", "gitconfig")
	if err := os.MkdirAll(filepath.Dir(existingPath), 0755); err != nil {
		t.Fatalf("mkdir existing source dir: %v", err)
	}
	original := "user-owned git config"
	if err := os.WriteFile(existingPath, []byte(original), 0644); err != nil {
		t.Fatalf("writing existing source: %v", err)
	}

	oldProfile := targetProfile
	oldFromRecipe := fromRecipe
	targetProfile = "default"
	fromRecipe = true
	defer func() {
		targetProfile = oldProfile
		fromRecipe = oldFromRecipe
	}()

	if err := runAdd(nil, []string{"git"}); err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}

	content, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("reading source after add: %v", err)
	}
	if string(content) != original {
		t.Fatalf("expected existing source to remain unchanged, got %q", string(content))
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
	removeUninstall = false
	removeYes = false
	removeDryRun = false
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

func TestRemoveAppWithUninstallUniquePackage(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// App definition with package + managed dotfile.
	bundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "tool",
		Package: &apps.Package{
			Apt: &apps.AptPackage{Name: "tool-pkg"},
		},
		Dotfiles: []apps.Dotfile{
			{Source: "tool/config", Target: "~/.toolrc"},
		},
	}
	if err := bundle.Save(filepath.Join(gdfDir, "apps", "tool.yaml")); err != nil {
		t.Fatalf("saving bundle: %v", err)
	}
	sourcePath := filepath.Join(gdfDir, "dotfiles", "tool", "config")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	targetPath := filepath.Join(homeDir, ".toolrc")
	if err := os.Symlink(sourcePath, targetPath); err != nil {
		t.Fatal(err)
	}

	targetProfile = "default"
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("loading profile: %v", err)
	}
	profile.Apps = append(profile.Apps, "tool")
	if err := profile.Save(profilePath); err != nil {
		t.Fatalf("saving profile: %v", err)
	}

	mockMgr := &MockPackageManager{mgrName: "apt"}
	oldOverride := packages.Override
	oldPlatform := platform.Override
	oldRemoveUninstall := removeUninstall
	oldRemoveYes := removeYes
	oldRemoveDryRun := removeDryRun
	packages.Override = mockMgr
	platform.Override = &platform.Platform{OS: "linux", Distro: "ubuntu", Home: homeDir}
	removeUninstall = true
	removeYes = true
	removeDryRun = false
	defer func() {
		packages.Override = oldOverride
		platform.Override = oldPlatform
		removeUninstall = oldRemoveUninstall
		removeYes = oldRemoveYes
		removeDryRun = oldRemoveDryRun
	}()

	if err := runRemove(nil, []string{"tool"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}

	profile, err = config.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("loading profile after remove: %v", err)
	}
	if contains(profile.Apps, "tool") {
		t.Fatalf("app still present in profile after remove")
	}
	if len(mockMgr.uninstalled) != 1 || mockMgr.uninstalled[0] != "tool-pkg" {
		t.Fatalf("unexpected uninstall calls: %#v", mockMgr.uninstalled)
	}
	if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
		t.Fatalf("expected symlink to be removed, got err=%v", err)
	}
}

func TestRemoveAppWithUninstallSkipsSharedPackage(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	writeBundle := func(name string) {
		bundle := &apps.Bundle{
			TypeMeta: schema.TypeMeta{Kind: "App/v1"},
			Name:     name,
			Package:  &apps.Package{Apt: &apps.AptPackage{Name: "shared-pkg"}},
		}
		if err := bundle.Save(filepath.Join(gdfDir, "apps", name+".yaml")); err != nil {
			t.Fatalf("saving bundle %s: %v", name, err)
		}
	}
	writeBundle("tool-a")
	writeBundle("tool-b")

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("loading profile: %v", err)
	}
	profile.Apps = []string{"tool-a", "tool-b"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatalf("saving profile: %v", err)
	}

	targetProfile = "default"
	mockMgr := &MockPackageManager{mgrName: "apt"}
	oldOverride := packages.Override
	oldPlatform := platform.Override
	oldRemoveUninstall := removeUninstall
	oldRemoveYes := removeYes
	oldRemoveDryRun := removeDryRun
	packages.Override = mockMgr
	platform.Override = &platform.Platform{OS: "linux", Distro: "ubuntu", Home: homeDir}
	removeUninstall = true
	removeYes = true
	removeDryRun = false
	defer func() {
		packages.Override = oldOverride
		platform.Override = oldPlatform
		removeUninstall = oldRemoveUninstall
		removeYes = oldRemoveYes
		removeDryRun = oldRemoveDryRun
	}()

	if err := runRemove(nil, []string{"tool-a"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}
	if len(mockMgr.uninstalled) != 0 {
		t.Fatalf("expected no uninstall for shared package, got %#v", mockMgr.uninstalled)
	}
}

func TestAddAppInteractiveIncludesRecipeDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)

	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	oldProfile := targetProfile
	oldFromRecipe := fromRecipe
	oldInteractive := addInteractive
	oldYes := globalYes
	defer func() {
		targetProfile = oldProfile
		fromRecipe = oldFromRecipe
		addInteractive = oldInteractive
		globalYes = oldYes
	}()

	targetProfile = "default"
	fromRecipe = true
	addInteractive = true
	globalYes = true

	if err := runAdd(nil, []string{"backend-dev"}); err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("loading profile: %v", err)
	}

	for _, dep := range []string{"backend-dev", "go", "docker", "kubectl", "terraform"} {
		if !contains(profile.Apps, dep) {
			t.Fatalf("expected %q in profile apps: %#v", dep, profile.Apps)
		}
	}
}

func TestRemoveAppWithUninstallDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	bundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "tool",
		Package:  &apps.Package{Apt: &apps.AptPackage{Name: "tool-pkg"}},
	}
	if err := bundle.Save(filepath.Join(gdfDir, "apps", "tool.yaml")); err != nil {
		t.Fatalf("saving bundle: %v", err)
	}

	targetProfile = "default"
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("loading profile: %v", err)
	}
	profile.Apps = append(profile.Apps, "tool")
	if err := profile.Save(profilePath); err != nil {
		t.Fatalf("saving profile: %v", err)
	}

	mockMgr := &MockPackageManager{mgrName: "apt"}
	oldOverride := packages.Override
	oldPlatform := platform.Override
	oldRemoveUninstall := removeUninstall
	oldRemoveYes := removeYes
	oldRemoveDryRun := removeDryRun
	packages.Override = mockMgr
	platform.Override = &platform.Platform{OS: "linux", Distro: "ubuntu", Home: homeDir}
	removeUninstall = true
	removeYes = true
	removeDryRun = true
	defer func() {
		packages.Override = oldOverride
		platform.Override = oldPlatform
		removeUninstall = oldRemoveUninstall
		removeYes = oldRemoveYes
		removeDryRun = oldRemoveDryRun
	}()

	if err := runRemove(nil, []string{"tool"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}
	profile, err = config.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("loading profile after dry-run: %v", err)
	}
	if !contains(profile.Apps, "tool") {
		t.Fatal("dry-run should not remove app from profile")
	}
	if len(mockMgr.uninstalled) != 0 {
		t.Fatalf("dry-run should not uninstall packages, got %#v", mockMgr.uninstalled)
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

func TestFindOrphanedApps(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	referencedBundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "referenced",
	}
	orphanBundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "orphan",
	}
	if err := referencedBundle.Save(filepath.Join(gdfDir, "apps", "referenced.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := orphanBundle.Save(filepath.Join(gdfDir, "apps", "orphan.yaml")); err != nil {
		t.Fatal(err)
	}

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = []string{"referenced"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	orphans, err := findOrphanedApps(gdfDir)
	if err != nil {
		t.Fatalf("findOrphanedApps() error = %v", err)
	}
	if len(orphans) != 1 || orphans[0].Name != "orphan" {
		t.Fatalf("unexpected orphans: %#v", orphans)
	}
}

func TestPruneOrphanedApps_Archive(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	orphanBundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "orphan",
	}
	if err := orphanBundle.Save(filepath.Join(gdfDir, "apps", "orphan.yaml")); err != nil {
		t.Fatal(err)
	}
	orphanDotfile := filepath.Join(gdfDir, "dotfiles", "orphan", "config")
	if err := os.MkdirAll(filepath.Dir(orphanDotfile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(orphanDotfile, []byte("cfg"), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := pruneOrphanedApps(gdfDir, false, false, true)
	if err != nil {
		t.Fatalf("pruneOrphanedApps() error = %v", err)
	}
	if report.Mode != "archive" || len(report.Pruned) != 1 || report.Pruned[0] != "orphan" {
		t.Fatalf("unexpected report: %#v", report)
	}
	if _, err := os.Stat(filepath.Join(gdfDir, "apps", "orphan.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected orphan app yaml to be archived")
	}
	if _, err := os.Stat(filepath.Join(gdfDir, "dotfiles", "orphan")); !os.IsNotExist(err) {
		t.Fatalf("expected orphan dotfiles to be archived")
	}
	if _, err := os.Stat(filepath.Join(report.ArchiveRoot, "apps", "orphan.yaml")); err != nil {
		t.Fatalf("expected archived app yaml: %v", err)
	}
	if _, err := os.Stat(filepath.Join(report.ArchiveRoot, "dotfiles", "orphan", "config")); err != nil {
		t.Fatalf("expected archived dotfile: %v", err)
	}
}

func TestPruneOrphanedApps_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	orphanBundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "orphan",
	}
	if err := orphanBundle.Save(filepath.Join(gdfDir, "apps", "orphan.yaml")); err != nil {
		t.Fatal(err)
	}
	orphanDotfile := filepath.Join(gdfDir, "dotfiles", "orphan", "config")
	if err := os.MkdirAll(filepath.Dir(orphanDotfile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(orphanDotfile, []byte("cfg"), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := pruneOrphanedApps(gdfDir, true, false, true)
	if err != nil {
		t.Fatalf("pruneOrphanedApps() error = %v", err)
	}
	if report.Mode != "delete" || len(report.Pruned) != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if _, err := os.Stat(filepath.Join(gdfDir, "apps", "orphan.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected orphan app yaml to be deleted")
	}
	if _, err := os.Stat(filepath.Join(gdfDir, "dotfiles", "orphan")); !os.IsNotExist(err) {
		t.Fatalf("expected orphan dotfiles to be deleted")
	}
}

func TestRemoveApp_PrintsDanglingCleanupGuidance(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	targetProfile = "default"
	removeUninstall = false
	removeYes = false
	removeDryRun = false
	if err := runAdd(nil, []string{"git"}); err != nil {
		t.Fatalf("setup: runAdd() error = %v", err)
	}

	out := captureStdout(t, func() {
		if err := runRemove(nil, []string{"git"}); err != nil {
			t.Fatalf("runRemove() error = %v", err)
		}
	})
	if !strings.Contains(out, "no longer referenced by any profile") {
		t.Fatalf("expected dangling guidance in output, got: %s", out)
	}
}
