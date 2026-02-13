package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile represents a profile definition (profiles/*/profile.yaml).
type Profile struct {
	// Name is the unique identifier for this profile (required).
	Name string `yaml:"name"`

	// Description provides a human-readable explanation.
	Description string `yaml:"description,omitempty"`

	// Includes lists other profiles to include.
	Includes []string `yaml:"includes,omitempty"`

	// Apps lists the app bundles in this profile.
	Apps []string `yaml:"apps,omitempty"`

	// Conditions defines conditional app inclusion/exclusion.
	Conditions []ProfileCondition `yaml:"conditions,omitempty"`
}

// ProfileCondition defines conditional app inclusion.
type ProfileCondition struct {
	// If is the condition expression (e.g., "os == 'macos'").
	If string `yaml:"if"`

	// IncludeApps lists apps to include when condition is true.
	IncludeApps []string `yaml:"include_apps,omitempty"`

	// ExcludeApps lists apps to exclude when condition is true.
	ExcludeApps []string `yaml:"exclude_apps,omitempty"`

	// Includes lists profiles to include when condition is true.
	Includes []string `yaml:"includes,omitempty"`
}

// LoadProfile reads a profile from a file.
func LoadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading profile file: %w", err)
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parsing profile YAML: %w", err)
	}

	return &profile, nil
}

// LoadProfileFromDir loads a profile from a profile directory.
// Looks for profile.yaml in the given directory.
func LoadProfileFromDir(dir string) (*Profile, error) {
	return LoadProfile(filepath.Join(dir, "profile.yaml"))
}

// LoadAllProfiles loads all profiles from a profiles directory.
// Each subdirectory should contain a profile.yaml file.
func LoadAllProfiles(profilesDir string) ([]*Profile, error) {
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading profiles directory: %w", err)
	}

	var profiles []*Profile
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		profilePath := filepath.Join(profilesDir, entry.Name(), "profile.yaml")
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			continue
		}

		profile, err := LoadProfile(profilePath)
		if err != nil {
			return nil, fmt.Errorf("loading profile %s: %w", entry.Name(), err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// Save writes the profile to a file.
func (p *Profile) Save(path string) error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshaling profile: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing profile file: %w", err)
	}

	return nil
}

// Validate checks that a profile is valid.
func (p *Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}
	return nil
}
