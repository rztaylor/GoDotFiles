package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rztaylor/GoDotFiles/internal/state"
)

func TestStatusCommand_NoState(t *testing.T) {
	tmpDir := t.TempDir()

	// Set HOME to tmpDir for the test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// No state file exists - should handle gracefully
	// We can't easily test the command output without refactoring,
	// but we can test the state loading logic
	st, err := state.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if len(st.AppliedProfiles) != 0 {
		t.Errorf("AppliedProfiles count = %d, want 0", len(st.AppliedProfiles))
	}
}

func TestStatusCommand_RequiresInitialization(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	var out bytes.Buffer
	statusCmd.SetOut(&out)
	defer statusCmd.SetOut(os.Stdout)

	err := runStatus(statusCmd, nil)
	if err == nil {
		t.Fatal("runStatus() expected error for uninitialized repo")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("runStatus() error = %v, want not initialized message", err)
	}
}

func TestStatusCommand_WithProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a state file with profiles
	st := &state.State{
		AppliedProfiles: []state.AppliedProfile{
			{
				Name:      "base",
				Apps:      []string{"git", "zsh", "vim"},
				AppliedAt: time.Now().Add(-2 * time.Hour),
			},
			{
				Name:      "work",
				Apps:      []string{"kubectl", "terraform"},
				AppliedAt: time.Now().Add(-1 * time.Hour),
			},
		},
		LastApplied: time.Now().Add(-1 * time.Hour),
	}

	statePath := filepath.Join(tmpDir, "state.yaml")
	if err := st.Save(statePath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load it back
	loaded, err := state.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if len(loaded.AppliedProfiles) != 2 {
		t.Errorf("AppliedProfiles count = %d, want 2", len(loaded.AppliedProfiles))
	}

	// Verify we can get all apps
	allApps := loaded.GetAppliedApps()
	if len(allApps) != 5 {
		t.Errorf("GetAppliedApps() count = %d, want 5", len(allApps))
	}
}

func TestStatusCommand_EmptyState(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty state file
	st := &state.State{
		AppliedProfiles: []state.AppliedProfile{},
	}

	statePath := filepath.Join(tmpDir, "state.yaml")
	if err := st.Save(statePath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load it back
	loaded, err := state.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if len(loaded.AppliedProfiles) != 0 {
		t.Errorf("AppliedProfiles count = %d, want 0", len(loaded.AppliedProfiles))
	}
}

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"just now", 30 * time.Second, "just now"},
		{"1 minute", 1 * time.Minute, "1 minute ago"},
		{"5 minutes", 5 * time.Minute, "5 minutes ago"},
		{"1 hour", 1 * time.Hour, "1 hour ago"},
		{"3 hours", 3 * time.Hour, "3 hours ago"},
		{"1 day", 24 * time.Hour, "1 day ago"},
		{"5 days", 5 * 24 * time.Hour, "5 days ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pastTime := time.Now().Add(-tt.duration)
			got := formatTimeAgo(pastTime)
			if got != tt.want {
				t.Errorf("formatTimeAgo() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{10, "s"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.count)), func(t *testing.T) {
			got := pluralize(tt.count)
			if got != tt.want {
				t.Errorf("pluralize(%d) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}
