package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/state"
	"github.com/rztaylor/GoDotFiles/internal/version"
)

const githubRepoSlug = "rztaylor/GoDotFiles"

var githubAPIBaseURL = "https://api.github.com"
var githubHTTPClient = &http.Client{Timeout: 5 * time.Second}

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
	latest, err := GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}

	// Compare versions
	// version.Version defaults to "0.6.0-dev" or similar, which is valid SemVer.
	currentVersion, err := semver.Parse(version.Version)
	if err != nil {
		// If current version is not valid semver (e.g. non-standard dev build), skip update check.
		return nil, nil
	}

	if latest == nil {
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

// GetLatestVersion fetches the latest version information from GitHub.
// It returns (nil, nil) when release info is unavailable.
func GetLatestVersion() (*selfupdate.Release, error) {
	baseURL := strings.TrimRight(githubAPIBaseURL, "/")
	url := fmt.Sprintf("%s/repos/%s/releases/latest", baseURL, githubRepoSlug)

	resp, err := githubHTTPClient.Get(url)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var release struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		HTMLURL     string `json:"html_url"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding github release: %w", err)
	}

	vStr := release.TagName
	if len(vStr) > 0 && vStr[0] == 'v' {
		vStr = vStr[1:]
	}
	v, err := semver.Parse(vStr)
	if err != nil {
		return nil, fmt.Errorf("parsing release version %q: %w", release.TagName, err)
	}

	osName := strings.ToLower(runtime.GOOS)
	archName := strings.ToLower(runtime.GOARCH)

	var assetURL string
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osName) && (strings.Contains(name, archName) || (archName == "amd64" && strings.Contains(name, "x86_64"))) {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}

	if assetURL == "" {
		return nil, nil
	}

	pubTime, _ := time.Parse(time.RFC3339, release.PublishedAt)

	return &selfupdate.Release{
		Version:       v,
		AssetURL:      assetURL,
		AssetByteSize: 0, // not needed for display
		AssetID:       0, // not needed
		URL:           release.HTMLURL,
		ReleaseNotes:  release.Body,
		Name:          release.Name,
		PublishedAt:   &pubTime,
		RepoName:      githubRepoSlug,
		RepoOwner:     strings.Split(githubRepoSlug, "/")[0],
	}, nil
}
