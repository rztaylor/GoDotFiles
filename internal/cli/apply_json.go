package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/engine"
	"github.com/rztaylor/GoDotFiles/internal/library"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

type applyDryRunPlan struct {
	Profiles []string             `json:"profiles"`
	Apps     []string             `json:"apps"`
	Risks    []engine.RiskFinding `json:"risks,omitempty"`
}

func runApplyDryRunJSON(cmd *cobra.Command, profileNames []string, gdfDir string, plat *platform.Platform, cfg *config.Config) error {
	profilesDir := filepath.Join(gdfDir, "profiles")
	allProfiles, err := config.LoadAllProfiles(profilesDir)
	if err != nil {
		return fmt.Errorf("loading profiles: %w", err)
	}

	profileMap := config.ProfileMap(allProfiles)
	resolvedProfiles, err := config.ResolveProfiles(profileNames, profileMap, plat)
	if err != nil {
		return fmt.Errorf("resolving profiles: %w", err)
	}

	appNames := make(map[string]bool)
	for _, profile := range resolvedProfiles {
		for _, app := range profile.Apps {
			appNames[app] = true
		}
	}

	appsDir := filepath.Join(gdfDir, "apps")
	allBundles := make(map[string]*apps.Bundle)
	var queue []string
	for appName := range appNames {
		queue = append(queue, appName)
	}

	libMgr := library.New()
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if _, exists := allBundles[name]; exists {
			continue
		}

		appPath := filepath.Join(appsDir, name+".yaml")
		bundle, err := apps.Load(appPath)
		if err != nil {
			if os.IsNotExist(err) {
				recipe, libErr := libMgr.Get(name)
				if libErr != nil {
					continue
				}
				bundle = recipe.ToBundle()
			} else {
				continue
			}
		}

		allBundles[name] = bundle
		queue = append(queue, bundle.Dependencies...)
	}

	appNamesSlice := make([]string, 0, len(allBundles))
	for name := range allBundles {
		appNamesSlice = append(appNamesSlice, name)
	}

	resolvedApps, err := apps.ResolveApps(appNamesSlice, allBundles)
	if err != nil {
		return fmt.Errorf("resolving app dependencies: %w", err)
	}

	plan := applyDryRunPlan{
		Profiles: make([]string, 0, len(resolvedProfiles)),
		Apps:     make([]string, 0, len(resolvedApps)),
		Risks:    engine.DetectHighRiskConfigurations(resolvedApps),
	}
	for _, p := range resolvedProfiles {
		plan.Profiles = append(plan.Profiles, p.Name)
	}
	for _, app := range resolvedApps {
		plan.Apps = append(plan.Apps, app.Name)
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(plan); err != nil {
		return fmt.Errorf("encoding dry-run JSON: %w", err)
	}

	confirmScripts := true
	if cfg.Security != nil {
		confirmScripts = cfg.Security.ConfirmScriptsDefault()
	}
	if len(plan.Risks) > 0 && !applyAllowRisky && confirmScripts {
		return withExitCode(fmt.Errorf("high-risk configuration detected in dry-run"), exitCodeHealthIssues)
	}

	return nil
}
