package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/git"
	"github.com/rztaylor/GoDotFiles/internal/library"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

func runHealthValidateReport(gdfDir string) (*healthReport, error) {
	report := &healthReport{
		Command: "health validate",
		OK:      true,
	}

	if !git.IsRepository(gdfDir) {
		report.add(healthFinding{
			Code:     "repo_not_initialized",
			Severity: healthSeverityError,
			Title:    "GDF repository is not initialized",
			Path:     gdfDir,
			Hint:     "Run 'gdf init' to create ~/.gdf",
		})
		report.sort()
		return report, nil
	}

	validateConfig(gdfDir, report)
	profiles, profileNames := validateProfiles(gdfDir, report)
	validateAppBundles(gdfDir, profiles, profileNames, report)
	report.sort()
	return report, nil
}

func validateConfig(gdfDir string, report *healthReport) {
	cfgPath := filepath.Join(gdfDir, "config.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		report.add(healthFinding{
			Code:     "config_missing",
			Severity: healthSeverityError,
			Title:    "Missing config.yaml",
			Path:     cfgPath,
			Hint:     "Re-run 'gdf init' or create config.yaml with kind: Config/v1",
		})
		return
	}
	if _, err := config.LoadConfig(cfgPath); err != nil {
		report.add(healthFinding{
			Code:     "config_invalid",
			Severity: healthSeverityError,
			Title:    "Invalid config.yaml",
			Path:     cfgPath,
			Detail:   err.Error(),
		})
	}
}

func validateProfiles(gdfDir string, report *healthReport) ([]*config.Profile, map[string]bool) {
	profilesDir := filepath.Join(gdfDir, "profiles")
	profiles, err := config.LoadAllProfiles(profilesDir)
	if err != nil {
		report.add(healthFinding{
			Code:     "profiles_load_failed",
			Severity: healthSeverityError,
			Title:    "Failed to load profiles",
			Path:     profilesDir,
			Detail:   err.Error(),
		})
		return nil, map[string]bool{}
	}
	if len(profiles) == 0 {
		report.add(healthFinding{
			Code:     "profiles_empty",
			Severity: healthSeverityWarning,
			Title:    "No profiles found",
			Path:     profilesDir,
			Hint:     "Create one with 'gdf profile create <name>'",
		})
	}

	profileNames := make(map[string]bool, len(profiles))
	plat := platform.Detect()
	for _, profile := range profiles {
		profileNames[profile.Name] = true
		profilePath := filepath.Join(profilesDir, profile.Name, "profile.yaml")
		if err := profile.Validate(); err != nil {
			report.add(healthFinding{
				Code:     "profile_invalid",
				Severity: healthSeverityError,
				Title:    fmt.Sprintf("Invalid profile: %s", profile.Name),
				Path:     profilePath,
				Detail:   err.Error(),
			})
		}

		for i, cond := range profile.Conditions {
			if _, err := config.EvaluateCondition(cond.If, plat); err != nil {
				report.add(healthFinding{
					Code:     "profile_condition_invalid",
					Severity: healthSeverityError,
					Title:    fmt.Sprintf("Invalid profile condition in %s", profile.Name),
					Path:     profilePath,
					Detail:   fmt.Sprintf("conditions[%d]: %v", i, err),
				})
			}
		}

		for _, include := range profile.Includes {
			if !profileNames[include] {
				// evaluated after map is complete in second pass
			}
		}
	}

	for _, profile := range profiles {
		profilePath := filepath.Join(profilesDir, profile.Name, "profile.yaml")
		for _, include := range profile.Includes {
			if !profileNames[include] {
				report.add(healthFinding{
					Code:     "profile_include_missing",
					Severity: healthSeverityError,
					Title:    fmt.Sprintf("Profile %s includes missing profile %s", profile.Name, include),
					Path:     profilePath,
					Hint:     "Create the included profile or remove it from includes",
				})
			}
		}
	}

	return profiles, profileNames
}

func validateAppBundles(gdfDir string, profiles []*config.Profile, _ map[string]bool, report *healthReport) {
	appsDir := filepath.Join(gdfDir, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil && !os.IsNotExist(err) {
		report.add(healthFinding{
			Code:     "apps_dir_unreadable",
			Severity: healthSeverityError,
			Title:    "Apps directory cannot be read",
			Path:     appsDir,
			Detail:   err.Error(),
		})
		return
	}

	localApps := make(map[string]bool)
	plat := platform.Detect()
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		appPath := filepath.Join(appsDir, entry.Name())
		bundle, err := apps.Load(appPath)
		if err != nil {
			report.add(healthFinding{
				Code:     "app_invalid",
				Severity: healthSeverityError,
				Title:    fmt.Sprintf("Invalid app bundle: %s", entry.Name()),
				Path:     appPath,
				Detail:   err.Error(),
			})
			continue
		}

		localApps[bundle.Name] = true
		if err := bundle.Validate(); err != nil {
			report.add(healthFinding{
				Code:     "app_semantic_invalid",
				Severity: healthSeverityError,
				Title:    fmt.Sprintf("App validation failed: %s", bundle.Name),
				Path:     appPath,
				Detail:   err.Error(),
			})
		}

		for i, dotfile := range bundle.Dotfiles {
			if dotfile.When == "" {
				continue
			}
			if _, err := config.EvaluateCondition(dotfile.When, plat); err != nil {
				report.add(healthFinding{
					Code:     "app_dotfile_condition_invalid",
					Severity: healthSeverityError,
					Title:    fmt.Sprintf("Invalid dotfile condition in app %s", bundle.Name),
					Path:     appPath,
					Detail:   fmt.Sprintf("dotfiles[%d]: %v", i, err),
				})
			}
		}
	}

	lib := library.New()
	for _, profile := range profiles {
		profilePath := filepath.Join(gdfDir, "profiles", profile.Name, "profile.yaml")
		for _, appName := range profile.Apps {
			if localApps[appName] {
				continue
			}
			if _, err := lib.Get(appName); err != nil {
				report.add(healthFinding{
					Code:     "profile_app_missing",
					Severity: healthSeverityError,
					Title:    fmt.Sprintf("Profile %s references unknown app %s", profile.Name, appName),
					Path:     profilePath,
					Hint:     "Create the app bundle locally or add a valid library recipe name",
				})
			}
		}
	}
}
