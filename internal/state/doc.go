// Package state manages the local state of applied profiles.
//
// The state package tracks which profiles have been applied to the current machine.
// This information is stored in ~/.gdf/state.yaml and is LOCAL ONLY (gitignored).
// State does not sync across machines.
//
// # Key Types
//
// State: The root state structure containing all applied profiles
// AppliedProfile: Represents a single profile that has been applied
//
// # Usage Example
//
//	st, err := state.LoadFromDir("~/.gdf")
//	if err != nil {
//	    return err
//	}
//
//	st.AddProfile("base", []string{"git", "zsh", "vim"})
//	if err := st.Save("~/.gdf/state.yaml"); err != nil {
//	    return err
//	}
//
//	if st.IsApplied("base") {
//	    fmt.Println("Base profile is applied")
//	}
//
// # Dependencies
//
// This package depends on:
//   - gopkg.in/yaml.v3 for YAML marshaling/unmarshaling
//   - Standard library for file I/O and time handling
package state
