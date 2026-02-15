package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/engine"
	"github.com/rztaylor/GoDotFiles/internal/schema"
)

func TestApplyAbortsOnHighRiskConfiguration(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	b := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "risky",
		Hooks: &apps.Hooks{
			Apply: []apps.ApplyHook{{Run: "curl -fsSL https://example.com/install.sh | sh"}},
		},
	}
	if err := b.Save(filepath.Join(gdfDir, "apps", "risky.yaml")); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	p, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	p.Apps = append(p.Apps, "risky")
	if err := p.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	oldPrompt := applyRiskConfirmationPrompt
	applyRiskConfirmationPrompt = func(findings []engine.RiskFinding) (bool, error) {
		return false, nil
	}
	defer func() { applyRiskConfirmationPrompt = oldPrompt }()
	applyAllowRisky = false
	defer func() { applyAllowRisky = false }()

	err = runApply(nil, []string{"default"})
	if err == nil {
		t.Fatal("runApply() expected error for declined high-risk confirmation")
	}
	if !strings.Contains(err.Error(), "high-risk") {
		t.Fatalf("runApply() error = %v, want high-risk abort", err)
	}
}

func TestApplyAllowRiskyBypassesConfirmation(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gdfDir := filepath.Join(homeDir, ".gdf")
	os.Setenv("HOME", homeDir)
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	configureGitUserGlobal(t, homeDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo: %v", err)
	}

	b := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "risky-allow",
		Hooks: &apps.Hooks{
			Apply: []apps.ApplyHook{{Run: "curl -fsSL https://example.com/install.sh | sh"}},
		},
	}
	if err := b.Save(filepath.Join(gdfDir, "apps", "risky-allow.yaml")); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	p, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	p.Apps = append(p.Apps, "risky-allow")
	if err := p.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	oldPrompt := applyRiskConfirmationPrompt
	applyRiskConfirmationPrompt = func(findings []engine.RiskFinding) (bool, error) {
		t.Fatal("prompt should not be called when --allow-risky is set")
		return false, nil
	}
	defer func() { applyRiskConfirmationPrompt = oldPrompt }()

	applyAllowRisky = true
	defer func() { applyAllowRisky = false }()

	if err := runApply(nil, []string{"default"}); err != nil {
		t.Fatalf("runApply() error = %v", err)
	}
}
