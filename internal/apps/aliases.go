package apps

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// GlobalAliases holds aliases that are not associated with any specific app bundle.
// These are stored in ~/.gdf/aliases.yaml.
type GlobalAliases struct {
	Aliases map[string]string `yaml:"aliases"`
}

// LoadGlobalAliases loads global aliases from a YAML file.
// Returns an empty GlobalAliases if the file does not exist.
func LoadGlobalAliases(path string) (*GlobalAliases, error) {
	ga := &GlobalAliases{
		Aliases: make(map[string]string),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ga, nil
		}
		return nil, fmt.Errorf("reading global aliases: %w", err)
	}

	if err := yaml.Unmarshal(data, ga); err != nil {
		return nil, fmt.Errorf("parsing global aliases: %w", err)
	}

	if ga.Aliases == nil {
		ga.Aliases = make(map[string]string)
	}

	return ga, nil
}

// Save writes global aliases to a YAML file.
func (ga *GlobalAliases) Save(path string) error {
	data, err := yaml.Marshal(ga)
	if err != nil {
		return fmt.Errorf("marshaling global aliases: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing global aliases: %w", err)
	}

	return nil
}

// Add adds or overwrites a global alias. Returns the previous value if it existed.
func (ga *GlobalAliases) Add(name, command string) (previous string, existed bool) {
	previous, existed = ga.Aliases[name]
	ga.Aliases[name] = command
	return
}

// Remove removes a global alias. Returns true if the alias existed.
func (ga *GlobalAliases) Remove(name string) bool {
	if _, ok := ga.Aliases[name]; !ok {
		return false
	}
	delete(ga.Aliases, name)
	return true
}

// SortedNames returns alias names in sorted order.
func (ga *GlobalAliases) SortedNames() []string {
	names := make([]string, 0, len(ga.Aliases))
	for name := range ga.Aliases {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
