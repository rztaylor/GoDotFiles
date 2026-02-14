package updater

import (
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/rztaylor/GoDotFiles/internal/version"
)

// Update performs the self-update to the latest version.
func Update() error {
	latest, found, err := selfupdate.DetectLatest("rztaylor/GoDotFiles")
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s is not found", "rztaylor/GoDotFiles")
	}

	// Parse current version
	vStr := version.Version
	if len(vStr) > 0 && vStr[0] == 'v' {
		vStr = vStr[1:]
	}
	currentVersion, err := semver.Parse(vStr)
	if err == nil {
		if latest.Version.LE(currentVersion) {
			fmt.Printf("Current version (%s) is the latest\n", version.Version)
			return nil
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable path: %w", err)
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}

	fmt.Printf("Successfully updated to %s\n", latest.Version)
	fmt.Println("Release note:")
	fmt.Println(latest.ReleaseNotes)
	return nil
}
