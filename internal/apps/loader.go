package apps

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads an app bundle from a YAML file.
func Load(path string) (*Bundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading bundle file: %w", err)
	}

	var bundle Bundle
	if err := yaml.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("parsing bundle YAML: %w", err)
	}

	return &bundle, nil
}

// LoadAll reads all app bundles from a directory.
// It looks for *.yaml files in the given directory.
func LoadAll(dir string) ([]*Bundle, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading apps directory: %w", err)
	}

	var bundles []*Bundle
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		bundle, err := Load(path)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", entry.Name(), err)
		}
		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

// Save writes an app bundle to a YAML file.
func (b *Bundle) Save(path string) error {
	data, err := yaml.Marshal(b)
	if err != nil {
		return fmt.Errorf("marshaling bundle: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing bundle file: %w", err)
	}

	return nil
}
