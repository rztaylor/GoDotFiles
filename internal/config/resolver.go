package config

import (
	"fmt"

	"github.com/rztaylor/GoDotFiles/internal/platform"
)

// ResolveProfiles resolves profile dependencies and returns an ordered list.
// The returned list is in dependency order (dependencies first).
// It evaluates profile conditions against the provided platform.
func ResolveProfiles(names []string, allProfiles map[string]*Profile, p *platform.Platform) ([]*Profile, error) {
	var result []*Profile
	visited := make(map[string]bool)
	inStack := make(map[string]bool) // For cycle detection

	var visit func(name string) error
	visit = func(name string) error {
		if inStack[name] {
			return fmt.Errorf("circular dependency detected: %s", name)
		}
		if visited[name] {
			return nil
		}

		originalProfile, ok := allProfiles[name]
		if !ok {
			return fmt.Errorf("profile not found: %s", name)
		}

		inStack[name] = true

		// Create effective profile (shallow copy to avoid mutating original)
		effectiveProfile := *originalProfile

		// Evaluate conditions
		condIncludes, condAddApps, condRemoveApps, err := CheckConditions(effectiveProfile.Conditions, p)
		if err != nil {
			return fmt.Errorf("evaluating conditions for profile %s: %w", name, err)
		}

		// Merge includes
		if len(condIncludes) > 0 {
			effectiveProfile.Includes = append(effectiveProfile.Includes, condIncludes...)
		}

		// Merge apps (Add)
		if len(condAddApps) > 0 {
			effectiveProfile.Apps = append(effectiveProfile.Apps, condAddApps...)
		}

		// Handle app exclusions
		if len(condRemoveApps) > 0 {
			effectiveProfile.Apps = filterExcluded(effectiveProfile.Apps, condRemoveApps)
		}

		// Visit dependencies first (including conditional ones)
		for _, include := range effectiveProfile.Includes {
			if err := visit(include); err != nil {
				return err
			}
		}

		inStack[name] = false
		visited[name] = true

		// Return the effective profile, not the original
		result = append(result, &effectiveProfile)

		return nil
	}

	for _, name := range names {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func filterExcluded(apps []string, excluded []string) []string {
	if len(excluded) == 0 {
		return apps
	}

	excludeMap := make(map[string]bool)
	for _, ex := range excluded {
		excludeMap[ex] = true
	}

	var result []string
	for _, app := range apps {
		if !excludeMap[app] {
			result = append(result, app)
		}
	}
	return result
}

// ProfileMap creates a map of profiles by name.
func ProfileMap(profiles []*Profile) map[string]*Profile {
	m := make(map[string]*Profile, len(profiles))
	for _, p := range profiles {
		m[p.Name] = p
	}
	return m
}
