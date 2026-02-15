package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/engine"
)

func TestRunAdd_NonInteractiveUsesDefaultPromptChoices(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	oldNonInteractive := globalNonInteractive
	oldYes := globalYes
	oldProfile := targetProfile
	oldFromRecipe := fromRecipe
	globalNonInteractive = true
	globalYes = false
	targetProfile = "default"
	fromRecipe = false
	defer func() {
		globalNonInteractive = oldNonInteractive
		globalYes = oldYes
		targetProfile = oldProfile
		fromRecipe = oldFromRecipe
	}()

	if err := runAdd(nil, []string{"kubectl"}); err != nil {
		t.Fatalf("runAdd() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(gdfDir, "apps", "kubectl.yaml")); err != nil {
		t.Fatalf("expected app bundle to be created: %v", err)
	}
}

func TestRunInstall_NonInteractiveNeedsExplicitPackageHint(t *testing.T) {
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

	oldNonInteractive := globalNonInteractive
	oldYes := globalYes
	oldInstallPkg := installPackage
	oldInstallProfile := installProfile
	globalNonInteractive = true
	globalYes = false
	installPackage = ""
	installProfile = "default"
	defer func() {
		globalNonInteractive = oldNonInteractive
		globalYes = oldYes
		installPackage = oldInstallPkg
		installProfile = oldInstallProfile
	}()

	err := runInstall(nil, []string{"needs-learning"})
	if err == nil {
		t.Fatal("expected non-interactive stop error, got nil")
	}
	if got := ExitCode(err); got != exitCodeNonInteractiveStop {
		t.Fatalf("ExitCode(err) = %d, want %d", got, exitCodeNonInteractiveStop)
	}
}

func TestRunRollback_NonInteractiveRequiresYesFlag(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	if err := os.MkdirAll(gdfDir, 0755); err != nil {
		t.Fatal(err)
	}

	logger := engine.NewLogger(false)
	logger.Log("link", filepath.Join(tmpDir, "target"), nil)
	if _, err := logger.Save(gdfDir); err != nil {
		t.Fatalf("logger.Save() error = %v", err)
	}

	oldNonInteractive := globalNonInteractive
	oldYes := globalYes
	oldRollbackYes := rollbackYes
	oldRollbackChoose := rollbackChooseSnapshot
	oldRollbackTarget := rollbackTarget
	globalNonInteractive = true
	globalYes = false
	rollbackYes = false
	rollbackChooseSnapshot = false
	rollbackTarget = ""
	defer func() {
		globalNonInteractive = oldNonInteractive
		globalYes = oldYes
		rollbackYes = oldRollbackYes
		rollbackChooseSnapshot = oldRollbackChoose
		rollbackTarget = oldRollbackTarget
	}()

	err := runRollback(nil, nil)
	if err == nil {
		t.Fatal("expected non-interactive stop error, got nil")
	}
	if got := ExitCode(err); got != exitCodeNonInteractiveStop {
		t.Fatalf("ExitCode(err) = %d, want %d", got, exitCodeNonInteractiveStop)
	}
}

func TestRunRestore_NonInteractiveRequiresInteractiveAck(t *testing.T) {
	oldNonInteractive := globalNonInteractive
	oldYes := globalYes
	globalNonInteractive = true
	globalYes = false
	defer func() {
		globalNonInteractive = oldNonInteractive
		globalYes = oldYes
	}()

	err := runRestore(nil, nil)
	if err == nil {
		t.Fatal("expected non-interactive stop error, got nil")
	}
	if got := ExitCode(err); got != exitCodeNonInteractiveStop {
		t.Fatalf("ExitCode(err) = %d, want %d", got, exitCodeNonInteractiveStop)
	}
}

func TestRunHealthFix_NonInteractiveRequiresExplicitApproval(t *testing.T) {
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

	for _, dir := range []string{"apps", "profiles", "dotfiles"} {
		if err := os.RemoveAll(filepath.Join(gdfDir, dir)); err != nil {
			t.Fatalf("removing %s: %v", dir, err)
		}
	}

	oldNonInteractive := globalNonInteractive
	oldYes := globalYes
	globalNonInteractive = true
	globalYes = false
	defer func() {
		globalNonInteractive = oldNonInteractive
		globalYes = oldYes
	}()

	var out bytes.Buffer
	err := runHealthFix(gdfDir, &out)
	if err == nil {
		t.Fatal("expected non-interactive stop error, got nil")
	}
	if got := ExitCode(err); got != exitCodeNonInteractiveStop {
		t.Fatalf("ExitCode(err) = %d, want %d", got, exitCodeNonInteractiveStop)
	}
}
