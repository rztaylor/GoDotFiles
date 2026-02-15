package cli

import (
	"path/filepath"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/schema"
)

func TestCollectAppStats_LocalBundle(t *testing.T) {
	tmpHome := t.TempDir()
	gdfDir := filepath.Join(tmpHome, ".gdf")
	t.Setenv("HOME", tmpHome)
	configureGitUserGlobal(t, tmpHome)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	bundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "git",
		Dotfiles: []apps.Dotfile{
			{Source: "git/config", Target: "~/.gitconfig"},
			{Source: "git/secret", Target: "~/.gitsecret", Secret: true},
		},
		Shell: &apps.Shell{
			Aliases: map[string]string{
				"g": "git",
			},
		},
	}
	if err := bundle.Save(filepath.Join(gdfDir, "apps", "git.yaml")); err != nil {
		t.Fatal(err)
	}

	stats := collectAppStats(gdfDir, []string{"git"})
	if len(stats) != 1 {
		t.Fatalf("len(stats) = %d, want 1", len(stats))
	}
	if stats[0].Source != "local" {
		t.Fatalf("Source = %q, want local", stats[0].Source)
	}
	if stats[0].Dotfiles != 2 {
		t.Fatalf("Dotfiles = %d, want 2", stats[0].Dotfiles)
	}
	if stats[0].Aliases != 1 {
		t.Fatalf("Aliases = %d, want 1", stats[0].Aliases)
	}
	if stats[0].Secrets != 1 {
		t.Fatalf("Secrets = %d, want 1", stats[0].Secrets)
	}
}

func TestCollectAppStats_MissingBundle(t *testing.T) {
	tmpHome := t.TempDir()
	gdfDir := filepath.Join(tmpHome, ".gdf")
	t.Setenv("HOME", tmpHome)
	configureGitUserGlobal(t, tmpHome)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	stats := collectAppStats(gdfDir, []string{"does-not-exist"})
	if len(stats) != 1 {
		t.Fatalf("len(stats) = %d, want 1", len(stats))
	}
	if stats[0].Source != "missing" {
		t.Fatalf("Source = %q, want missing", stats[0].Source)
	}
}

func TestAggregateAppStats(t *testing.T) {
	dotfiles, aliases, secrets := aggregateAppStats([]appStats{
		{Name: "a", Dotfiles: 2, Aliases: 1, Secrets: 1},
		{Name: "b", Dotfiles: 3, Aliases: 4, Secrets: 0},
	})
	if dotfiles != 5 || aliases != 5 || secrets != 1 {
		t.Fatalf("aggregateAppStats() = (%d,%d,%d), want (5,5,1)", dotfiles, aliases, secrets)
	}
}
