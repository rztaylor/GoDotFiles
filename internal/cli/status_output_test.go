package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/schema"
	"github.com/rztaylor/GoDotFiles/internal/state"
)

func TestCollectStatusReportIncludesDrift(t *testing.T) {
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

	bundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "git",
		Dotfiles: []apps.Dotfile{
			{Source: "git/config", Target: "~/.config/gdf-test/git-config"},
		},
	}
	if err := bundle.Save(filepath.Join(gdfDir, "apps", "git.yaml")); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(gdfDir, "dotfiles", "git", "config")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = []string{"git"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	st := &state.State{}
	st.AddProfile("default", []string{"git"})
	if err := st.Save(filepath.Join(gdfDir, "state.yaml")); err != nil {
		t.Fatal(err)
	}

	report, err := collectStatusReport(gdfDir, driftOptions{IncludeIssues: true, IncludePreview: true})
	if err != nil {
		t.Fatalf("collectStatusReport() error = %v", err)
	}
	if report.Drift.TargetMissing == 0 {
		t.Fatalf("expected target_missing drift issue")
	}
	if len(report.Drift.Issues) == 0 {
		t.Fatalf("expected drift issues in detailed report")
	}
}

func TestStatusJSONOutput(t *testing.T) {
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

	st := &state.State{}
	if err := st.Save(filepath.Join(gdfDir, "state.yaml")); err != nil {
		t.Fatal(err)
	}

	oldJSON := statusJSON
	statusJSON = true
	defer func() { statusJSON = oldJSON }()

	var out bytes.Buffer
	statusCmd.SetOut(&out)
	defer statusCmd.SetOut(os.Stdout)

	if err := runStatus(statusCmd, nil); err != nil {
		t.Fatalf("runStatus() error = %v", err)
	}
	if !strings.Contains(out.String(), "\"drift\"") {
		t.Fatalf("expected JSON status output, got: %s", out.String())
	}
}

func TestCollectStatusReportPatchOutputAndCache(t *testing.T) {
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

	bundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "git",
		Dotfiles: []apps.Dotfile{
			{Source: "git/config", Target: "~/.config/gdf-test/git-config"},
		},
	}
	if err := bundle.Save(filepath.Join(gdfDir, "apps", "git.yaml")); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(gdfDir, "dotfiles", "git", "config")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("new-value\n"), 0644); err != nil {
		t.Fatal(err)
	}
	targetPath := filepath.Join(homeDir, ".config", "gdf-test", "git-config")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetPath, []byte("old-value\n"), 0644); err != nil {
		t.Fatal(err)
	}

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = []string{"git"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	st := &state.State{}
	st.AddProfile("default", []string{"git"})
	if err := st.Save(filepath.Join(gdfDir, "state.yaml")); err != nil {
		t.Fatal(err)
	}

	report, err := collectStatusReport(gdfDir, driftOptions{
		IncludeIssues:  true,
		IncludePreview: true,
		IncludePatch:   true,
		PatchMaxBytes:  1024 * 1024,
		PatchMaxFiles:  20,
	})
	if err != nil {
		t.Fatalf("collectStatusReport() error = %v", err)
	}
	if len(report.Drift.Issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(report.Drift.Issues))
	}
	issue := report.Drift.Issues[0]
	if issue.Patch == "" {
		t.Fatalf("expected patch output, got empty issue: %#v", issue)
	}
	if issue.PatchSkippedReason != "" {
		t.Fatalf("expected no patch skip reason, got %q", issue.PatchSkippedReason)
	}
	if _, err := os.Stat(filepath.Join(gdfDir, ".cache", "status-diff.yaml")); err != nil {
		t.Fatalf("expected status diff cache file to be written: %v", err)
	}
}

func TestCollectStatusReportPatchRespectsMaxBytes(t *testing.T) {
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

	bundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "git",
		Dotfiles: []apps.Dotfile{
			{Source: "git/config", Target: "~/.config/gdf-test/git-config"},
		},
	}
	if err := bundle.Save(filepath.Join(gdfDir, "apps", "git.yaml")); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(gdfDir, "dotfiles", "git", "config")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("1234567890"), 0644); err != nil {
		t.Fatal(err)
	}
	targetPath := filepath.Join(homeDir, ".config", "gdf-test", "git-config")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetPath, []byte("abcdefghij"), 0644); err != nil {
		t.Fatal(err)
	}

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = []string{"git"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	st := &state.State{}
	st.AddProfile("default", []string{"git"})
	if err := st.Save(filepath.Join(gdfDir, "state.yaml")); err != nil {
		t.Fatal(err)
	}

	report, err := collectStatusReport(gdfDir, driftOptions{
		IncludeIssues:  true,
		IncludePreview: true,
		IncludePatch:   true,
		PatchMaxBytes:  4,
		PatchMaxFiles:  20,
	})
	if err != nil {
		t.Fatalf("collectStatusReport() error = %v", err)
	}
	if len(report.Drift.Issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(report.Drift.Issues))
	}
	issue := report.Drift.Issues[0]
	if issue.Patch != "" {
		t.Fatalf("expected no patch due to max-bytes, got: %s", issue.Patch)
	}
	if !strings.Contains(issue.PatchSkippedReason, "max-bytes") {
		t.Fatalf("expected max-bytes skip reason, got %q", issue.PatchSkippedReason)
	}
}
