package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

func TestCoreWorkflowRegression_InitAddTrackApplyStatusRollback(t *testing.T) {
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

	oldTargetProfile := targetProfile
	oldFromRecipe := fromRecipe
	targetProfile = "default"
	fromRecipe = true
	defer func() {
		targetProfile = oldTargetProfile
		fromRecipe = oldFromRecipe
	}()

	if err := runAdd(nil, []string{"git"}); err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}

	trackPath := filepath.Join(homeDir, ".gdf-regression")
	if err := os.WriteFile(trackPath, []byte("regression"), 0644); err != nil {
		t.Fatalf("writing tracked file: %v", err)
	}

	oldTargetApp := targetApp
	oldSecret := secretFlag
	targetApp = "git"
	secretFlag = false
	defer func() {
		targetApp = oldTargetApp
		secretFlag = oldSecret
	}()

	if err := runTrack(nil, []string{trackPath}); err != nil {
		t.Fatalf("runTrack() error = %v", err)
	}

	oldApplyDryRun := applyDryRun
	oldApplyAllowRisky := applyAllowRisky
	oldPkgOverride := packages.Override
	oldPlatformOverride := platform.Override
	applyDryRun = false
	applyAllowRisky = true
	packages.Override = &MockPackageManager{mgrName: "apt"}
	platform.Override = &platform.Platform{OS: "linux", Distro: "ubuntu", Home: homeDir}
	defer func() {
		applyDryRun = oldApplyDryRun
		applyAllowRisky = oldApplyAllowRisky
		packages.Override = oldPkgOverride
		platform.Override = oldPlatformOverride
	}()

	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply() error = %v", err)
	}

	report, err := collectStatusReport(gdfDir, driftOptions{})
	if err != nil {
		t.Fatalf("collectStatusReport() error = %v", err)
	}
	if len(report.AppliedProfiles) == 0 {
		t.Fatal("expected at least one applied profile after apply")
	}

	oldRollbackYes := rollbackYes
	oldRollbackChoose := rollbackChooseSnapshot
	oldRollbackTarget := rollbackTarget
	rollbackYes = true
	rollbackChooseSnapshot = false
	rollbackTarget = ""
	defer func() {
		rollbackYes = oldRollbackYes
		rollbackChooseSnapshot = oldRollbackChoose
		rollbackTarget = oldRollbackTarget
	}()

	if err := runRollback(nil, nil); err != nil {
		t.Fatalf("runRollback() error = %v", err)
	}
}
