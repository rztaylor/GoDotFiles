package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/config"
)

func TestParseCSVAppNames(t *testing.T) {
	got := parseCSVAppNames(" git, kubectl ,git,,AWS CLI ")
	want := []string{"git", "kubectl", "aws-cli"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRunInitSetup_NonInteractiveJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configureGitUserGlobal(t, home)

	oldNonInteractive := globalNonInteractive
	oldSetupProfile := setupProfile
	oldSetupApps := setupApps
	oldSetupJSON := setupJSON
	defer func() {
		globalNonInteractive = oldNonInteractive
		setupProfile = oldSetupProfile
		setupApps = oldSetupApps
		setupJSON = oldSetupJSON
	}()

	globalNonInteractive = true
	setupProfile = "work"
	setupApps = "git,kubectl"
	setupJSON = true

	outPath := filepath.Join(home, "out.txt")
	f, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = f

	err = runInitSetup(nil, nil)

	_ = f.Close()
	os.Stdout = oldStdout
	if err != nil {
		t.Fatalf("runInitSetup() error = %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	idx := 0
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] == '{' {
			idx = i
			break
		}
	}
	var summary initSetupSummary
	if err := json.Unmarshal(data[idx:], &summary); err != nil {
		t.Fatalf("unmarshal setup output: %v; output=%s", err, string(data))
	}
	if summary.Profile != "work" {
		t.Fatalf("profile = %q, want work", summary.Profile)
	}
	if len(summary.Apps) != 2 {
		t.Fatalf("apps len = %d, want 2", len(summary.Apps))
	}

	profilePath := filepath.Join(home, ".gdf", "profiles", "work", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("loading profile: %v", err)
	}
	if !contains(profile.Apps, "git") || !contains(profile.Apps, "kubectl") {
		t.Fatalf("profile apps = %#v, want git/kubectl", profile.Apps)
	}
}
