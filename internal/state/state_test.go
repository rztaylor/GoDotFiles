package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadState(t *testing.T) {
	tests := []struct {
		name             string
		yaml             string
		wantProfileCount int
		wantErr          bool
	}{
		{
			name:             "empty state file returns empty state",
			yaml:             "",
			wantProfileCount: 0,
		},
		{
			name: "valid state with one profile",
			yaml: `applied_profiles:
  - name: base
    apps:
      - git
      - zsh
    applied_at: 2024-02-11T18:30:00Z
last_applied: 2024-02-11T18:30:00Z
`,
			wantProfileCount: 1,
		},
		{
			name: "valid state with multiple profiles",
			yaml: `applied_profiles:
  - name: base
    apps:
      - git
    applied_at: 2024-02-11T18:30:00Z
  - name: work
    apps:
      - kubectl
    applied_at: 2024-02-11T18:31:00Z
last_applied: 2024-02-11T18:31:00Z
`,
			wantProfileCount: 2,
		},
		{
			name:    "invalid yaml",
			yaml:    `applied_profiles: [invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "state.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("writing temp file: %v", err)
			}

			st, err := Load(path)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if len(st.AppliedProfiles) != tt.wantProfileCount {
				t.Errorf("AppliedProfiles count = %d, want %d", len(st.AppliedProfiles), tt.wantProfileCount)
			}
		})
	}
}

func TestLoadState_NonExistent(t *testing.T) {
	st, err := Load("/nonexistent/state.yaml")
	if err != nil {
		t.Fatalf("Load() error = %v, want nil for missing file", err)
	}
	if st == nil {
		t.Error("Load() returned nil, want empty state")
	}
	if len(st.AppliedProfiles) != 0 {
		t.Errorf("AppliedProfiles count = %d, want 0", len(st.AppliedProfiles))
	}
}

func TestSaveState(t *testing.T) {
	st := &State{
		AppliedProfiles: []AppliedProfile{
			{
				Name:      "base",
				Apps:      []string{"git", "zsh"},
				AppliedAt: time.Date(2024, 2, 11, 18, 30, 0, 0, time.UTC),
			},
		},
		LastApplied: time.Date(2024, 2, 11, 18, 30, 0, 0, time.UTC),
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "state.yaml")

	if err := st.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load it back and verify
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.AppliedProfiles) != 1 {
		t.Errorf("AppliedProfiles count = %d, want 1", len(loaded.AppliedProfiles))
	}

	if loaded.AppliedProfiles[0].Name != "base" {
		t.Errorf("Profile name = %q, want %q", loaded.AppliedProfiles[0].Name, "base")
	}

	if len(loaded.AppliedProfiles[0].Apps) != 2 {
		t.Errorf("Apps count = %d, want 2", len(loaded.AppliedProfiles[0].Apps))
	}
}

func TestAddProfile(t *testing.T) {
	tests := []struct {
		name            string
		initialProfiles []AppliedProfile
		addName         string
		addApps         []string
		wantCount       int
		wantApps        []string
	}{
		{
			name:            "add to empty state",
			initialProfiles: nil,
			addName:         "base",
			addApps:         []string{"git", "zsh"},
			wantCount:       1,
			wantApps:        []string{"git", "zsh"},
		},
		{
			name: "update existing profile",
			initialProfiles: []AppliedProfile{
				{
					Name:      "base",
					Apps:      []string{"git"},
					AppliedAt: time.Date(2024, 2, 11, 18, 0, 0, 0, time.UTC),
				},
			},
			addName:   "base",
			addApps:   []string{"git", "zsh", "vim"},
			wantCount: 1,
			wantApps:  []string{"git", "zsh", "vim"},
		},
		{
			name: "add new profile to existing state",
			initialProfiles: []AppliedProfile{
				{
					Name:      "base",
					Apps:      []string{"git"},
					AppliedAt: time.Date(2024, 2, 11, 18, 0, 0, 0, time.UTC),
				},
			},
			addName:   "work",
			addApps:   []string{"kubectl"},
			wantCount: 2,
			wantApps:  []string{"kubectl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &State{
				AppliedProfiles: tt.initialProfiles,
			}

			st.AddProfile(tt.addName, tt.addApps)

			if len(st.AppliedProfiles) != tt.wantCount {
				t.Errorf("AppliedProfiles count = %d, want %d", len(st.AppliedProfiles), tt.wantCount)
			}

			// Find the added/updated profile
			var found *AppliedProfile
			for i := range st.AppliedProfiles {
				if st.AppliedProfiles[i].Name == tt.addName {
					found = &st.AppliedProfiles[i]
					break
				}
			}

			if found == nil {
				t.Fatalf("profile %q not found", tt.addName)
			}

			if len(found.Apps) != len(tt.wantApps) {
				t.Errorf("Apps count = %d, want %d", len(found.Apps), len(tt.wantApps))
			}

			for i, app := range tt.wantApps {
				if found.Apps[i] != app {
					t.Errorf("Apps[%d] = %q, want %q", i, found.Apps[i], app)
				}
			}

			// Verify AppliedAt is recent
			if time.Since(found.AppliedAt) > time.Minute {
				t.Errorf("AppliedAt is not recent: %v", found.AppliedAt)
			}

			// Verify LastApplied is updated
			if time.Since(st.LastApplied) > time.Minute {
				t.Errorf("LastApplied is not recent: %v", st.LastApplied)
			}
		})
	}
}

func TestRemoveProfile(t *testing.T) {
	tests := []struct {
		name            string
		initialProfiles []AppliedProfile
		removeName      string
		wantCount       int
	}{
		{
			name:            "remove from empty state",
			initialProfiles: nil,
			removeName:      "base",
			wantCount:       0,
		},
		{
			name: "remove existing profile",
			initialProfiles: []AppliedProfile{
				{Name: "base", Apps: []string{"git"}},
				{Name: "work", Apps: []string{"kubectl"}},
			},
			removeName: "base",
			wantCount:  1,
		},
		{
			name: "remove non-existent profile",
			initialProfiles: []AppliedProfile{
				{Name: "base", Apps: []string{"git"}},
			},
			removeName: "nonexistent",
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &State{
				AppliedProfiles: tt.initialProfiles,
			}

			st.RemoveProfile(tt.removeName)

			if len(st.AppliedProfiles) != tt.wantCount {
				t.Errorf("AppliedProfiles count = %d, want %d", len(st.AppliedProfiles), tt.wantCount)
			}

			// Verify the removed profile is gone
			for _, p := range st.AppliedProfiles {
				if p.Name == tt.removeName {
					t.Errorf("profile %q still exists after removal", tt.removeName)
				}
			}
		})
	}
}

func TestIsApplied(t *testing.T) {
	st := &State{
		AppliedProfiles: []AppliedProfile{
			{Name: "base", Apps: []string{"git"}},
			{Name: "work", Apps: []string{"kubectl"}},
		},
	}

	tests := []struct {
		name    string
		profile string
		want    bool
	}{
		{"existing profile", "base", true},
		{"another existing profile", "work", true},
		{"non-existent profile", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := st.IsApplied(tt.profile); got != tt.want {
				t.Errorf("IsApplied(%q) = %v, want %v", tt.profile, got, tt.want)
			}
		})
	}
}

func TestGetAppliedApps(t *testing.T) {
	tests := []struct {
		name     string
		profiles []AppliedProfile
		want     []string
	}{
		{
			name:     "empty state",
			profiles: nil,
			want:     []string{},
		},
		{
			name: "single profile",
			profiles: []AppliedProfile{
				{Name: "base", Apps: []string{"git", "zsh"}},
			},
			want: []string{"git", "zsh"},
		},
		{
			name: "multiple profiles with unique apps",
			profiles: []AppliedProfile{
				{Name: "base", Apps: []string{"git", "zsh"}},
				{Name: "work", Apps: []string{"kubectl", "terraform"}},
			},
			want: []string{"git", "zsh", "kubectl", "terraform"},
		},
		{
			name: "multiple profiles with duplicate apps",
			profiles: []AppliedProfile{
				{Name: "base", Apps: []string{"git", "zsh"}},
				{Name: "work", Apps: []string{"git", "kubectl"}},
			},
			want: []string{"git", "zsh", "kubectl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &State{
				AppliedProfiles: tt.profiles,
			}

			got := st.GetAppliedApps()

			if len(got) != len(tt.want) {
				t.Errorf("GetAppliedApps() count = %d, want %d", len(got), len(tt.want))
			}

			// Convert to map for easy lookup (order doesn't matter)
			gotMap := make(map[string]bool)
			for _, app := range got {
				gotMap[app] = true
			}

			for _, app := range tt.want {
				if !gotMap[app] {
					t.Errorf("GetAppliedApps() missing app %q", app)
				}
			}
		})
	}
}

func TestLoadFromDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a state file in the directory
	st := &State{
		AppliedProfiles: []AppliedProfile{
			{Name: "base", Apps: []string{"git"}},
		},
	}

	statePath := filepath.Join(tmpDir, "state.yaml")
	if err := st.Save(statePath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load from directory
	loaded, err := LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if len(loaded.AppliedProfiles) != 1 {
		t.Errorf("AppliedProfiles count = %d, want 1", len(loaded.AppliedProfiles))
	}

	if loaded.AppliedProfiles[0].Name != "base" {
		t.Errorf("Profile name = %q, want %q", loaded.AppliedProfiles[0].Name, "base")
	}
}
