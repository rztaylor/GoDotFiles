package cli

import (
	"bytes"
	"os"
	"path/filepath"
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
	globalYes = true
	globalNonInteractive = false
	defer func() {
		globalYes = oldYes
		globalNonInteractive = oldNonInteractive
	}()

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
