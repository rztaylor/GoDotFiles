package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHealthValidateReport_UninitializedRepo(t *testing.T) {
	tmpDir := t.TempDir()
	report, err := runHealthValidateReport(filepath.Join(tmpDir, ".gdf"))
	if err != nil {
		t.Fatalf("runHealthValidateReport() error = %v", err)
	}
	if report.Errors == 0 {
		t.Fatalf("expected validation errors for missing repo, got 0")
	}
	if report.Findings[0].Code != "repo_not_initialized" {
		t.Fatalf("first finding code = %s, want repo_not_initialized", report.Findings[0].Code)
	}
}

func TestHealthDoctorAndFix_CreateMissingDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	for _, dir := range []string{"apps", "profiles", "dotfiles"} {
		if err := os.RemoveAll(filepath.Join(gdfDir, dir)); err != nil {
			t.Fatalf("removing %s: %v", dir, err)
		}
	}

	doctor, err := runHealthDoctorReport(gdfDir)
	if err != nil {
		t.Fatalf("runHealthDoctorReport() error = %v", err)
	}
	if doctor.Errors == 0 {
		t.Fatalf("expected doctor to report blocking issues")
	}

	oldYes := globalYes
	oldNonInteractive := globalNonInteractive
	oldGuarded := healthFixGuarded
	oldDryRun := healthFixDryRun
	globalYes = true
	globalNonInteractive = false
	defer func() {
		globalYes = oldYes
		globalNonInteractive = oldNonInteractive
		healthFixGuarded = oldGuarded
		healthFixDryRun = oldDryRun
	}()
	healthFixGuarded = false
	healthFixDryRun = false

	var out bytes.Buffer
	if err := runHealthFix(gdfDir, &out); err != nil {
		t.Fatalf("runHealthFix() error = %v", err)
	}

	for _, dir := range []string{"apps", "profiles", "dotfiles"} {
		if _, err := os.Stat(filepath.Join(gdfDir, dir)); err != nil {
			t.Fatalf("expected %s to be recreated: %v", dir, err)
		}
	}
}

func TestHealthFixGuardedRepairsInvalidConfigWithBackup(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	configPath := filepath.Join(gdfDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("kind: Config/v1\n: invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	oldYes := globalYes
	oldNonInteractive := globalNonInteractive
	oldGuarded := healthFixGuarded
	oldDryRun := healthFixDryRun
	globalYes = true
	globalNonInteractive = false
	healthFixGuarded = true
	healthFixDryRun = false
	defer func() {
		globalYes = oldYes
		globalNonInteractive = oldNonInteractive
		healthFixGuarded = oldGuarded
		healthFixDryRun = oldDryRun
	}()

	var out bytes.Buffer
	if err := runHealthFix(gdfDir, &out); err != nil {
		t.Fatalf("runHealthFix() error = %v", err)
	}

	matches, err := filepath.Glob(configPath + ".gdf.bak.*")
	if err != nil {
		t.Fatalf("glob backup files: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("expected backup file for guarded config repair")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "kind: Config/v1") {
		t.Fatalf("expected rewritten config with valid kind, got: %s", string(content))
	}
}

func TestHealthFixDryRunDoesNotMutate(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	if err := os.RemoveAll(filepath.Join(gdfDir, "apps")); err != nil {
		t.Fatal(err)
	}

	oldYes := globalYes
	oldNonInteractive := globalNonInteractive
	oldGuarded := healthFixGuarded
	oldDryRun := healthFixDryRun
	globalYes = true
	globalNonInteractive = false
	healthFixGuarded = false
	healthFixDryRun = true
	defer func() {
		globalYes = oldYes
		globalNonInteractive = oldNonInteractive
		healthFixGuarded = oldGuarded
		healthFixDryRun = oldDryRun
	}()

	var out bytes.Buffer
	if err := runHealthFix(gdfDir, &out); err != nil {
		t.Fatalf("runHealthFix() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(gdfDir, "apps")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not recreate missing apps dir")
	}
	if !strings.Contains(out.String(), "Dry run only") {
		t.Fatalf("expected dry-run output, got: %s", out.String())
	}
}

func TestHealthCIExitCodeOnErrors(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	err := runHealthCI(healthCICmd, nil)
	if err == nil {
		t.Fatalf("runHealthCI() expected error for missing repo")
	}
	if got := ExitCode(err); got != exitCodeHealthIssues {
		t.Fatalf("ExitCode(err) = %d, want %d", got, exitCodeHealthIssues)
	}
}
