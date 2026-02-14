package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// State represents the local state of applied profiles.
type State struct {
	// AppliedProfiles is the list of profiles that have been applied.
	AppliedProfiles []AppliedProfile `yaml:"applied_profiles"`

	// LastApplied is the timestamp of the last apply operation.
	LastApplied time.Time `yaml:"last_applied"`

	// UpdateCheck tracks the last update check.
	UpdateCheck UpdateCheck `yaml:"update_check"`

	// Path is the file path where state is stored (not serialized).
	Path string `yaml:"-"`
}

// UpdateCheck tracks the state of auto-update checks.
type UpdateCheck struct {
	// LastChecked is when the last check was performed.
	LastChecked time.Time `yaml:"last_checked"`

	// SnoozeUntil is the time until which checks are snoozed.
	SnoozeUntil time.Time `yaml:"snooze_until"`
}

// AppliedProfile represents a single profile that has been applied.
type AppliedProfile struct {
	// Name is the profile name.
	Name string `yaml:"name"`

	// Apps is the list of apps in this profile.
	Apps []string `yaml:"apps"`

	// AppliedAt is when this profile was applied.
	AppliedAt time.Time `yaml:"applied_at"`
}

// Load reads the state from a file.
// If the file doesn't exist, returns an empty state.
func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty state if file doesn't exist
			return &State{
				AppliedProfiles: []AppliedProfile{},
				Path:            path,
			}, nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var st State
	if err := yaml.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("parsing state YAML: %w", err)
	}

	// Ensure AppliedProfiles is not nil
	if st.AppliedProfiles == nil {
		st.AppliedProfiles = []AppliedProfile{}
	}

	st.Path = path
	return &st, nil
}

// LoadFromDir loads state.yaml from a directory.
func LoadFromDir(dir string) (*State, error) {
	return Load(filepath.Join(dir, "state.yaml"))
}

// Save writes the state to a file.
func (s *State) Save(path string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}

// AddProfile adds or updates a profile in the state.
func (s *State) AddProfile(name string, apps []string) {
	now := time.Now()

	// Check if profile already exists
	for i := range s.AppliedProfiles {
		if s.AppliedProfiles[i].Name == name {
			// Update existing profile
			s.AppliedProfiles[i].Apps = apps
			s.AppliedProfiles[i].AppliedAt = now
			s.LastApplied = now
			return
		}
	}

	// Add new profile
	s.AppliedProfiles = append(s.AppliedProfiles, AppliedProfile{
		Name:      name,
		Apps:      apps,
		AppliedAt: now,
	})
	s.LastApplied = now
}

// RemoveProfile removes a profile from the state.
func (s *State) RemoveProfile(name string) {
	for i := range s.AppliedProfiles {
		if s.AppliedProfiles[i].Name == name {
			// Remove by slicing
			s.AppliedProfiles = append(s.AppliedProfiles[:i], s.AppliedProfiles[i+1:]...)
			return
		}
	}
}

// IsApplied checks if a profile is currently applied.
func (s *State) IsApplied(profileName string) bool {
	for _, p := range s.AppliedProfiles {
		if p.Name == profileName {
			return true
		}
	}
	return false
}

// GetAppliedApps returns a deduplicated list of all apps from all applied profiles.
func (s *State) GetAppliedApps() []string {
	appSet := make(map[string]bool)
	for _, profile := range s.AppliedProfiles {
		for _, app := range profile.Apps {
			appSet[app] = true
		}
	}

	apps := make([]string, 0, len(appSet))
	for app := range appSet {
		apps = append(apps, app)
	}

	return apps
}
