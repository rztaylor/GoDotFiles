package apps

import (
	"fmt"
)

// ResolveApps resolves app bundle dependencies and returns an ordered list.
// The returned list is in dependency order (dependencies first).
// This performs a topological sort on the app dependency graph.
func ResolveApps(names []string, allBundles map[string]*Bundle) ([]*Bundle, error) {
	var result []*Bundle
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

		bundle, ok := allBundles[name]
		if !ok {
			return fmt.Errorf("app not found: %s", name)
		}

		inStack[name] = true

		// Visit dependencies first
		for _, dep := range bundle.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}

		inStack[name] = false
		visited[name] = true
		result = append(result, bundle)

		return nil
	}

	for _, name := range names {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// BundleMap creates a map of bundles by name.
func BundleMap(bundles []*Bundle) map[string]*Bundle {
	m := make(map[string]*Bundle, len(bundles))
	for _, b := range bundles {
		m[b.Name] = b
	}
	return m
}
