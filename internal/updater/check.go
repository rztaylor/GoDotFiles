package updater

import (
	"fmt"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/state"
	"github.com/rztaylor/GoDotFiles/internal/version"
)

// ReleaseInfo holds information about a release.
type ReleaseInfo struct {
	Version     string
	PublishedAt time.Time
	URL         string
	AssetURL    string
}

// CheckForUpdate checks if a new version is available.
// It returns nil if no update is available or if checks are disabled/snoozed (unless force is true).
func CheckForUpdate(cfg *config.Config, st *state.State, force bool) (*ReleaseInfo, error) {
	if !force {
		// 1. Check if disabled
		if cfg.Updates != nil && cfg.Updates.Disabled {
			return nil, nil
		}

		// 2. Check if snoozed
		if !st.UpdateCheck.SnoozeUntil.IsZero() && time.Now().Before(st.UpdateCheck.SnoozeUntil) {
			return nil, nil
		}

		// 3. Check frequency (default 24h)
		interval := 24 * time.Hour
		if cfg.Updates != nil && cfg.Updates.CheckInterval != nil {
			interval = *cfg.Updates.CheckInterval
		}

		if time.Since(st.UpdateCheck.LastChecked) < interval {
			return nil, nil
		}
	}

	// Update LastChecked immediately to prevent spamming GitHub API
	st.UpdateCheck.LastChecked = time.Now()
	// Ignore save errors here, it's not critical
	_ = st.Save(st.Path)

	// 4. Perform check
	latest, found, err := selfupdate.DetectLatest("rztaylor/GoDotFiles")
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}

	if !found {
		return nil, nil
	}

	// Compare versions
	// version.Version defaults to "0.6.0-dev" or similar, which is valid SemVer.
	currentVersion, err := semver.Parse(version.Version)
	if err != nil {
		// If current version is not valid semver (e.g. non-standard dev build), skip update check.
		return nil, nil
	}

	if latest.Version.LE(currentVersion) {
		return nil, nil
	}

	pubTime := time.Time{}
	if latest.PublishedAt != nil {
		pubTime = *latest.PublishedAt
	}

	return &ReleaseInfo{
		Version:     latest.Version.String(),
		PublishedAt: pubTime,
		URL:         latest.URL,
		AssetURL:    latest.AssetURL,
	}, nil
}
