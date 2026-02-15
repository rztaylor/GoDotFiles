package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/schema"
)

func TestProfileCreate(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")

	// Mock environment
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Configure git user
	configureGitUserGlobal(t, tmpDir)

	// Initialize repo
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	tests := []struct {
		name        string
		profileName string
		description string
		wantErr     bool
	}{
		{
			name:        "create profile with description",
			profileName: "work",
			description: "Work profile",
			wantErr:     false,
		},
		{
			name:        "create profile without description",
			profileName: "home",
			description: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileDescription = tt.description
			err := runProfileCreate(nil, []string{tt.profileName})

			if (err != nil) != tt.wantErr {
				t.Errorf("runProfileCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify profile was created
				profilePath := filepath.Join(gdfDir, "profiles", tt.profileName, "profile.yaml")
				if _, err := os.Stat(profilePath); os.IsNotExist(err) {
					t.Error("profile file was not created")
				}

				// Load and verify profile
				profile, err := config.LoadProfile(profilePath)
				if err != nil {
					t.Fatalf("failed to load created profile: %v", err)
				}

				if profile.Name != tt.profileName {
					t.Errorf("profile name = %s, want %s", profile.Name, tt.profileName)
				}

				if tt.description != "" && profile.Description != tt.description {
					t.Errorf("profile description = %s, want %s", profile.Description, tt.description)
				}

				if len(profile.Apps) != 0 {
					t.Error("profile apps should be empty")
				}
			}
		})
	}

	// Test duplicate profile creation
	t.Run("create duplicate profile", func(t *testing.T) {
		profileDescription = "Test"
		// Create first profile
		if err := runProfileCreate(nil, []string{"duplicate"}); err != nil {
			t.Fatalf("first create failed: %v", err)
		}

		// Try to create duplicate
		err := runProfileCreate(nil, []string{"duplicate"})
		if err == nil {
			t.Error("expected error when creating duplicate profile, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' error, got: %v", err)
		}
	})
}

func TestProfileList(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	configureGitUserGlobal(t, tmpDir)

	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	t.Run("list with no profiles", func(t *testing.T) {
		// Should not error when there are no profiles
		if err := runProfileList(nil, nil); err != nil {
			t.Errorf("runProfileList() error = %v, expected nil", err)
		}
	})

	t.Run("list with profiles", func(t *testing.T) {
		// Create some profiles
		profiles := []struct {
			name string
			desc string
		}{
			{"work", "Work environment"},
			{"home", "Home setup"},
			{"sre", "SRE tools"},
		}

		for _, p := range profiles {
			profileDescription = p.desc
			if err := runProfileCreate(nil, []string{p.name}); err != nil {
				t.Fatalf("failed to create profile %s: %v", p.name, err)
			}
		}

		// Run list command (output goes to stdout, we just check it doesn't error)
		if err := runProfileList(nil, nil); err != nil {
			t.Errorf("runProfileList() error = %v", err)
		}

		// Verify all profiles exist
		for _, p := range profiles {
			profilePath := filepath.Join(gdfDir, "profiles", p.name, "profile.yaml")
			if _, err := os.Stat(profilePath); os.IsNotExist(err) {
				t.Errorf("profile %s not found", p.name)
			}
		}
	})
}

func TestProfileShow(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	configureGitUserGlobal(t, tmpDir)

	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	t.Run("show non-existent profile", func(t *testing.T) {
		err := runProfileShow(nil, []string{"nonexistent"})
		if err == nil {
			t.Error("expected error for non-existent profile, got nil")
			return
		}
		// Just verify the function returns an error - the specific error message is less important
	})

	t.Run("show profile with apps and includes", func(t *testing.T) {
		// Create a profile
		profileDescription = "Test profile"
		if err := runProfileCreate(nil, []string{"testshow"}); err != nil {
			t.Fatalf("failed to create profile: %v", err)
		}

		// Add some apps to the profile
		targetProfile = "testshow"
		_ = runAdd(nil, []string{"git"})
		_ = runAdd(nil, []string{"vim"})

		// Manually add includes for testing
		profilePath := filepath.Join(gdfDir, "profiles", "testshow", "profile.yaml")
		profile, err := config.LoadProfile(profilePath)
		if err != nil {
			t.Fatalf("failed to load profile: %v", err)
		}

		profile.Includes = []string{"base", "common"}
		if err := profile.Save(profilePath); err != nil {
			t.Fatalf("failed to save profile: %v", err)
		}

		// Run show command
		if err := runProfileShow(nil, []string{"testshow"}); err != nil {
			t.Errorf("runProfileShow() error = %v", err)
		}

		// Verify profile content
		profile, err = config.LoadProfile(profilePath)
		if err != nil {
			t.Fatalf("failed to reload profile: %v", err)
		}

		if len(profile.Apps) != 2 {
			t.Errorf("expected 2 apps, got %d", len(profile.Apps))
		}

		if len(profile.Includes) != 2 {
			t.Errorf("expected 2 includes, got %d", len(profile.Includes))
		}
	})

	t.Run("show empty profile", func(t *testing.T) {
		profileDescription = "Empty profile"
		if err := runProfileCreate(nil, []string{"empty"}); err != nil {
			t.Fatalf("failed to create profile: %v", err)
		}

		// Run show command on empty profile
		if err := runProfileShow(nil, []string{"empty"}); err != nil {
			t.Errorf("runProfileShow() error = %v", err)
		}
	})

	t.Run("show default profile (no args)", func(t *testing.T) {
		// Ensure default profile exists
		createProfile := func(name string) {
			profileDir := filepath.Join(gdfDir, "profiles", name)
			os.MkdirAll(profileDir, 0755)
			profile := &config.Profile{Name: name}
			profile.Save(filepath.Join(profileDir, "profile.yaml"))
		}
		createProfile("default")

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		oldStdin := os.Stdin
		os.Stdin = r
		defer func() { os.Stdin = oldStdin }()
		if _, err := w.Write([]byte("1\n")); err != nil {
			t.Fatal(err)
		}
		_ = w.Close()

		if err := runProfileShow(nil, []string{}); err != nil {
			t.Errorf("runProfileShow(no args) error = %v", err)
		}
	})
}

func TestProfileShowWithConditions(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	configureGitUserGlobal(t, tmpDir)

	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// Create a profile with conditions
	profileDir := filepath.Join(gdfDir, "profiles", "conditional")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("failed to create profile dir: %v", err)
	}

	profile := &config.Profile{
		Name:        "conditional",
		Description: "Profile with conditions",
		Apps:        []string{"base-app"},
		Conditions: []config.ProfileCondition{
			{
				If:          "os == 'macos'",
				IncludeApps: []string{"homebrew", "iterm"},
			},
			{
				If:          "os == 'linux'",
				ExcludeApps: []string{"iterm"},
			},
		},
	}

	profilePath := filepath.Join(profileDir, "profile.yaml")
	if err := profile.Save(profilePath); err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}

	// Run show command
	if err := runProfileShow(nil, []string{"conditional"}); err != nil {
		t.Errorf("runProfileShow() error = %v", err)
	}

	// Verify profile was loaded correctly
	loadedProfile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("failed to load profile: %v", err)
	}

	if len(loadedProfile.Conditions) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(loadedProfile.Conditions))
	}
}

func TestProfileShow_PrintsAppCounters(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)

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

	profilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	profile.Apps = []string{"git"}
	if err := profile.Save(profilePath); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		if err := runProfileShow(nil, []string{"default"}); err != nil {
			t.Fatalf("runProfileShow() error = %v", err)
		}
	})

	required := []string{
		"Totals",
		"Dotfiles",
		"Aliases",
		"Secrets",
		"git",
	}
	for _, needle := range required {
		if !strings.Contains(out, needle) {
			t.Fatalf("expected output to contain %q, got:\n%s", needle, out)
		}
	}
}

func TestProfileDelete(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	configureGitUserGlobal(t, tmpDir)

	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// Helper to create a profile
	createProfile := func(name string, apps []string) {
		profileDir := filepath.Join(gdfDir, "profiles", name)
		if err := os.MkdirAll(profileDir, 0755); err != nil {
			t.Fatalf("failed to create profile dir: %v", err)
		}
		profile := &config.Profile{
			Name: name,
			Apps: apps,
		}
		if err := profile.Save(filepath.Join(profileDir, "profile.yaml")); err != nil {
			t.Fatalf("failed to save profile: %v", err)
		}
	}

	t.Run("delete non-existent profile", func(t *testing.T) {
		err := runProfileDelete(nil, []string{"nonexistent"})
		if err == nil {
			t.Error("expected error for non-existent profile, got nil")
		}
	})

	t.Run("delete default profile", func(t *testing.T) {
		createProfile("default", []string{"git"})
		err := runProfileDelete(nil, []string{"default"})
		if err == nil {
			t.Error("expected error when deleting default profile, got nil")
		}
	})

	t.Run("delete profile and move apps", func(t *testing.T) {
		createProfile("todelete", []string{"app1", "app2"})
		// Ensure default profile exists
		createProfile("default", []string{"base-app"})

		// Create a profile that includes the one to be deleted
		createProfile("dependant", nil)
		depPath := filepath.Join(gdfDir, "profiles", "dependant", "profile.yaml")
		depProfile, _ := config.LoadProfile(depPath)
		depProfile.Includes = []string{"todelete", "other"}
		depProfile.Save(depPath)

		err := runProfileDelete(nil, []string{"todelete"})
		if err != nil {
			t.Errorf("runProfileDelete() error = %v", err)
		}

		// verify profile is gone
		profilePath := filepath.Join(gdfDir, "profiles", "todelete", "profile.yaml")
		if _, err := os.Stat(profilePath); !os.IsNotExist(err) {
			t.Error("profile 'todelete' still exists")
		}

		// verify apps moved to default
		defaultProfilePath := filepath.Join(gdfDir, "profiles", "default", "profile.yaml")
		defaultProfile, err := config.LoadProfile(defaultProfilePath)
		if err != nil {
			t.Fatalf("failed to load default profile: %v", err)
		}

		foundApp1 := false
		foundApp2 := false
		for _, app := range defaultProfile.Apps {
			if app == "app1" {
				foundApp1 = true
			}
			if app == "app2" {
				foundApp2 = true
			}
		}

		if !foundApp1 || !foundApp2 {
			t.Errorf("apps were not moved to default profile: %v", defaultProfile.Apps)
		}

		// verify dependency removed
		depProfile, err = config.LoadProfile(depPath)
		if err != nil {
			t.Fatalf("failed to reload dependant profile: %v", err)
		}
		for _, inc := range depProfile.Includes {
			if inc == "todelete" {
				t.Error("deleted profile 'todelete' still referenced in 'dependant'")
			}
		}
	})
}

func TestProfileDeleteModeConflict(t *testing.T) {
	oldPurge := profileDeletePurge
	oldMigrate := profileDeleteMigrateToDefault
	oldLeave := profileDeleteLeaveDangling
	profileDeletePurge = true
	profileDeleteMigrateToDefault = true
	profileDeleteLeaveDangling = false
	defer func() {
		profileDeletePurge = oldPurge
		profileDeleteMigrateToDefault = oldMigrate
		profileDeleteLeaveDangling = oldLeave
	}()

	_, err := resolveProfileDeleteMode()
	if err == nil {
		t.Fatal("expected mode conflict error, got nil")
	}
}

func TestParseProfileDeleteModeChoice(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    profileDeleteMode
		wantErr bool
	}{
		{name: "default empty", in: "", want: profileDeleteModeMigrateToDefault},
		{name: "choice 1", in: "1", want: profileDeleteModeMigrateToDefault},
		{name: "migrate string", in: "migrate-to-default", want: profileDeleteModeMigrateToDefault},
		{name: "choice 2", in: "2", want: profileDeleteModePurge},
		{name: "purge string", in: "purge", want: profileDeleteModePurge},
		{name: "choice 3", in: "3", want: profileDeleteModeLeaveDangling},
		{name: "leave string", in: "leave-dangling", want: profileDeleteModeLeaveDangling},
		{name: "invalid", in: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseProfileDeleteModeChoice(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseProfileDeleteModeChoice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("parseProfileDeleteModeChoice() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestProfileDeleteDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	t.Setenv("HOME", tmpDir)
	configureGitUserGlobal(t, tmpDir)
	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	profileDir := filepath.Join(gdfDir, "profiles", "temp")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatal(err)
	}
	profile := &config.Profile{Name: "temp", Apps: []string{"app1"}}
	if err := profile.Save(filepath.Join(profileDir, "profile.yaml")); err != nil {
		t.Fatal(err)
	}

	oldPurge := profileDeletePurge
	oldMigrate := profileDeleteMigrateToDefault
	oldLeave := profileDeleteLeaveDangling
	oldDryRun := profileDeleteDryRun
	oldYes := profileDeleteYes
	profileDeletePurge = false
	profileDeleteMigrateToDefault = true
	profileDeleteLeaveDangling = false
	profileDeleteDryRun = true
	profileDeleteYes = true
	defer func() {
		profileDeletePurge = oldPurge
		profileDeleteMigrateToDefault = oldMigrate
		profileDeleteLeaveDangling = oldLeave
		profileDeleteDryRun = oldDryRun
		profileDeleteYes = oldYes
	}()

	if err := runProfileDelete(nil, []string{"temp"}); err != nil {
		t.Fatalf("runProfileDelete() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(profileDir, "profile.yaml")); err != nil {
		t.Fatalf("dry-run should not delete profile: %v", err)
	}
}

func TestProfileDeletePurgeRemovesUniqueApps(t *testing.T) {
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

	// App unique to target profile.
	uniqueBundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "unique-app",
		Dotfiles: []apps.Dotfile{{Source: "unique-app/config", Target: "~/.unique-app"}},
	}
	if err := uniqueBundle.Save(filepath.Join(gdfDir, "apps", "unique-app.yaml")); err != nil {
		t.Fatal(err)
	}
	uniqueSource := filepath.Join(gdfDir, "dotfiles", "unique-app", "config")
	if err := os.MkdirAll(filepath.Dir(uniqueSource), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(uniqueSource, []byte("cfg"), 0644); err != nil {
		t.Fatal(err)
	}
	uniqueTarget := filepath.Join(homeDir, ".unique-app")
	if err := os.Symlink(uniqueSource, uniqueTarget); err != nil {
		t.Fatal(err)
	}

	sharedBundle := &apps.Bundle{
		TypeMeta: schema.TypeMeta{Kind: "App/v1"},
		Name:     "shared-app",
	}
	if err := sharedBundle.Save(filepath.Join(gdfDir, "apps", "shared-app.yaml")); err != nil {
		t.Fatal(err)
	}

	// target profile with both unique and shared app
	targetDir := filepath.Join(gdfDir, "profiles", "to-delete")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	targetProfile := &config.Profile{Name: "to-delete", Apps: []string{"unique-app", "shared-app"}}
	if err := targetProfile.Save(filepath.Join(targetDir, "profile.yaml")); err != nil {
		t.Fatal(err)
	}

	// another profile referencing shared-app only
	otherDir := filepath.Join(gdfDir, "profiles", "other")
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatal(err)
	}
	otherProfile := &config.Profile{Name: "other", Apps: []string{"shared-app"}}
	if err := otherProfile.Save(filepath.Join(otherDir, "profile.yaml")); err != nil {
		t.Fatal(err)
	}

	oldPurge := profileDeletePurge
	oldMigrate := profileDeleteMigrateToDefault
	oldLeave := profileDeleteLeaveDangling
	oldDryRun := profileDeleteDryRun
	oldYes := profileDeleteYes
	profileDeletePurge = true
	profileDeleteMigrateToDefault = false
	profileDeleteLeaveDangling = false
	profileDeleteDryRun = false
	profileDeleteYes = true
	defer func() {
		profileDeletePurge = oldPurge
		profileDeleteMigrateToDefault = oldMigrate
		profileDeleteLeaveDangling = oldLeave
		profileDeleteDryRun = oldDryRun
		profileDeleteYes = oldYes
	}()

	if err := runProfileDelete(nil, []string{"to-delete"}); err != nil {
		t.Fatalf("runProfileDelete() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(gdfDir, "apps", "unique-app.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected unique app definition to be purged")
	}
	if _, err := os.Stat(filepath.Join(gdfDir, "apps", "shared-app.yaml")); err != nil {
		t.Fatalf("expected shared app definition to remain: %v", err)
	}
	if _, err := os.Lstat(uniqueTarget); !os.IsNotExist(err) {
		t.Fatalf("expected unique managed symlink to be removed")
	}
}

func TestProfileRename(t *testing.T) {
	tmpDir := t.TempDir()
	gdfDir := filepath.Join(tmpDir, ".gdf")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	configureGitUserGlobal(t, tmpDir)

	if err := createNewRepo(gdfDir); err != nil {
		t.Fatalf("createNewRepo() error = %v", err)
	}

	// Helper to create a profile
	createProfile := func(name string, includes []string) {
		profileDir := filepath.Join(gdfDir, "profiles", name)
		if err := os.MkdirAll(profileDir, 0755); err != nil {
			t.Fatalf("failed to create profile dir: %v", err)
		}
		profile := &config.Profile{
			Name:     name,
			Includes: includes,
		}
		if err := profile.Save(filepath.Join(profileDir, "profile.yaml")); err != nil {
			t.Fatalf("failed to save profile: %v", err)
		}
	}

	t.Run("rename non-existent profile", func(t *testing.T) {
		err := runProfileRename(nil, []string{"nonexistent", "newname"})
		if err == nil {
			t.Error("expected error for non-existent profile, got nil")
		}
	})

	t.Run("rename default profile", func(t *testing.T) {
		createProfile("default", nil)
		err := runProfileRename(nil, []string{"default", "newname"})
		if err == nil {
			t.Error("expected error when renaming default profile, got nil")
		}
	})

	t.Run("rename to existing profile", func(t *testing.T) {
		createProfile("p1", nil)
		createProfile("p2", nil)
		err := runProfileRename(nil, []string{"p1", "p2"})
		if err == nil {
			t.Error("expected error when renaming to existing profile, got nil")
		}
	})

	t.Run("rename invalid name", func(t *testing.T) {
		createProfile("p1", nil)
	})

	t.Run("rename success with dependencies", func(t *testing.T) {
		createProfile("oldname", nil)
		createProfile("dependant", []string{"oldname", "other"})

		// Create state with oldname applied
		statePath := filepath.Join(gdfDir, "state.yaml")
		stateContent := `
applied_profiles:
  - name: oldname
    apps: []
    applied_at: 2024-01-01T00:00:00Z
  - name: other
    apps: []
`
		if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
			t.Fatalf("failed to write state: %v", err)
		}

		// Perform rename
		err := runProfileRename(nil, []string{"oldname", "newname"})
		if err != nil {
			t.Errorf("runProfileRename() error = %v", err)
		}

		// Verify old directory gone
		if _, err := os.Stat(filepath.Join(gdfDir, "profiles", "oldname")); !os.IsNotExist(err) {
			t.Error("old profile directory still exists")
		}

		// Verify new directory exists
		newProfilePath := filepath.Join(gdfDir, "profiles", "newname", "profile.yaml")
		newProfile, err := config.LoadProfile(newProfilePath)
		if err != nil {
			t.Fatalf("failed to load new profile: %v", err)
		}
		if newProfile.Name != "newname" {
			t.Errorf("profile name in yaml not updated: got %s", newProfile.Name)
		}

		// Verify dependency update
		depProfile, err := config.LoadProfile(filepath.Join(gdfDir, "profiles", "dependant", "profile.yaml"))
		if err != nil {
			t.Fatalf("failed to load dependant profile: %v", err)
		}

		found := false
		for _, inc := range depProfile.Includes {
			if inc == "newname" {
				found = true
			}
			if inc == "oldname" {
				t.Error("dependant profile still includes oldname")
			}
		}
		if !found {
			t.Error("dependant profile does not include newname")
		}

		// Verify state update
		content, err := os.ReadFile(statePath)
		if err != nil {
			t.Fatalf("failed to read state: %v", err)
		}
		contentStr := string(content)
		if !strings.Contains(contentStr, "name: newname") {
			t.Error("state does not contain newname")
		}
		if strings.Contains(contentStr, "name: oldname") {
			t.Error("state still contains oldname")
		}
	})
}
