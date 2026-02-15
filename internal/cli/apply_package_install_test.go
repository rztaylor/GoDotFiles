package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

func TestApplyPackageInstallBehavior(t *testing.T) {
	t.Run("skips install when package is already installed", func(t *testing.T) {
		_, _ = setupApplyPackageInstallTest(t, "pkg-app", "git")
		mgr := &MockPackageManager{
			mgrName:   "apt",
			installed: []string{"git"},
		}
		packages.Override = mgr
		defer func() { packages.Override = nil }()

		if err := runApply(nil, []string{"default"}); err != nil {
			t.Fatalf("runApply() error = %v", err)
		}
		if mgr.installCalls != 0 {
			t.Fatalf("installCalls = %d, want 0", mgr.installCalls)
		}
	})

	t.Run("installs when package is not installed", func(t *testing.T) {
		_, _ = setupApplyPackageInstallTest(t, "pkg-app", "git")
		mgr := &MockPackageManager{
			mgrName: "apt",
		}
		packages.Override = mgr
		defer func() { packages.Override = nil }()

		if err := runApply(nil, []string{"default"}); err != nil {
			t.Fatalf("runApply() error = %v", err)
		}
		if mgr.installCalls != 1 {
			t.Fatalf("installCalls = %d, want 1", mgr.installCalls)
		}
	})

	t.Run("installs when installed check returns error", func(t *testing.T) {
		_, _ = setupApplyPackageInstallTest(t, "pkg-app", "git")
		mgr := &MockPackageManager{
			mgrName:        "apt",
			isInstalledErr: errors.New("probe failed"),
		}
		packages.Override = mgr
		defer func() { packages.Override = nil }()

		if err := runApply(nil, []string{"default"}); err != nil {
			t.Fatalf("runApply() error = %v", err)
		}
		if mgr.installCalls != 1 {
			t.Fatalf("installCalls = %d, want 1", mgr.installCalls)
		}
	})

	t.Run("skips preferred manager install when package is installed via alternate manager", func(t *testing.T) {
		_, _ = setupApplyPackageInstallBundleTest(t, "pkg-app", &apps.Package{
			Apt:  &apps.AptPackage{Name: "git"},
			Brew: "git",
		})

		aptMgr := &MockPackageManager{mgrName: "apt"}
		brewMgr := &MockPackageManager{mgrName: "brew", installed: []string{"git"}}

		oldFactory := packageManagerFactory
		oldAuto := packageAutoManagerForPlatform
		packageManagerFactory = func(name string) (packages.Manager, bool) {
			switch name {
			case "apt":
				return aptMgr, true
			case "brew":
				return brewMgr, true
			default:
				return nil, false
			}
		}
		packageAutoManagerForPlatform = func(_ *platform.Platform) packages.Manager {
			return &MockPackageManager{mgrName: "apt"}
		}
		defer func() {
			packageManagerFactory = oldFactory
			packageAutoManagerForPlatform = oldAuto
		}()

		if err := runApply(nil, []string{"default"}); err != nil {
			t.Fatalf("runApply() error = %v", err)
		}
		if aptMgr.installCalls != 0 {
			t.Fatalf("apt installCalls = %d, want 0", aptMgr.installCalls)
		}
		if brewMgr.installCalls != 0 {
			t.Fatalf("brew installCalls = %d, want 0", brewMgr.installCalls)
		}
		if aptMgr.isInstalledCalls == 0 || brewMgr.isInstalledCalls == 0 {
			t.Fatalf("expected installed checks for both managers, got apt=%d brew=%d", aptMgr.isInstalledCalls, brewMgr.isInstalledCalls)
		}
	})

	t.Run("treats custom install as valid and skips execution during apply", func(t *testing.T) {
		tmpDir := t.TempDir()
		marker := filepath.Join(tmpDir, "custom-installed")
		confirm := false
		_, _ = setupApplyPackageInstallBundleTest(t, "pkg-app", &apps.Package{
			Custom: &apps.CustomInstall{
				Script:  "printf custom > " + marker,
				Confirm: &confirm,
			},
		})

		if err := runApply(nil, []string{"default"}); err != nil {
			t.Fatalf("runApply() error = %v", err)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("custom install script should not execute during apply")
		}
	})
}

func setupApplyPackageInstallTest(t *testing.T, appName, pkgName string) (string, string) {
	return setupApplyPackageInstallBundleTest(t, appName, &apps.Package{
		Apt: &apps.AptPackage{Name: pkgName},
	})
}

func setupApplyPackageInstallBundleTest(t *testing.T, appName string, pkg *apps.Package) (string, string) {
	t.Helper()

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

	platform.Override = &platform.Platform{OS: "linux", Distro: "ubuntu", Home: homeDir}
	t.Cleanup(func() { platform.Override = nil })

	oldDryRun := applyDryRun
	oldJSON := applyJSON
	oldAllowRisky := applyAllowRisky
	applyDryRun = false
	applyJSON = false
	applyAllowRisky = false
	t.Cleanup(func() {
		applyDryRun = oldDryRun
		applyJSON = oldJSON
		applyAllowRisky = oldAllowRisky
	})

	appPath := filepath.Join(gdfDir, "apps", appName+".yaml")
	bundle := &apps.Bundle{
		Name:    appName,
		Package: pkg,
	}
	if err := bundle.Save(appPath); err != nil {
		t.Fatalf("saving app bundle: %v", err)
	}

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("loading profile: %v", err)
	}
	profile.Apps = []string{appName}
	if err := profile.Save(profilePath); err != nil {
		t.Fatalf("saving profile: %v", err)
	}

	return homeDir, gdfDir
}
