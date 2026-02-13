// Package git provides Git repository operations for GDF.
//
// # Responsibility
//
// This package handles:
//   - Initializing new repositories
//   - Cloning existing repositories
//   - Staging and committing changes
//   - Push/pull operations
//   - Status checking
//
// # Key Types
//
//   - Repository: Represents a git repository
//   - Remote: Remote repository configuration
//
// # Usage
//
//	repo, err := git.Open("~/.gdf")
//	if err != nil {
//	    return err
//	}
//	repo.Add(".")
//	repo.Commit("Updated dotfiles")
//	repo.Push()
package git
