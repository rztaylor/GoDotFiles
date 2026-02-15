package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <app-pattern>",
	Short: "Move apps between profiles",
	Long: `Move apps matching a pattern from one profile to another.
	
At least one of --from or --to must be specified. 
If one is omitted, GDF selects the profile automatically when possible.`,
	Args: cobra.ExactArgs(1),
	RunE: runMove,
}

var (
	moveFromProfile string
	moveToProfile   string
)

func init() {
	appCmd.AddCommand(moveCmd)
	moveCmd.Flags().StringVar(&moveFromProfile, "from", "", "Source profile")
	moveCmd.Flags().StringVar(&moveToProfile, "to", "", "Target profile")
	moveCmd.Flags().BoolVar(&moveApply, "apply", false, "Preview and apply affected profiles after moving apps (requires confirmation unless --yes)")
}

func runMove(cmd *cobra.Command, args []string) error {
	pattern := args[0]
	gdfDir := platform.ConfigDir()

	// 1. Validation & Defaults
	from := moveFromProfile
	to := moveToProfile

	if from == "" && to == "" {
		return fmt.Errorf("must specify at least one of --from or --to")
	}

	if from == "" {
		resolvedFrom, err := resolveProfileSelectionForCommand(gdfDir, from, "gdf app move --from")
		if err != nil {
			return err
		}
		from = resolvedFrom
	}
	if to == "" {
		resolvedTo, err := resolveProfileSelectionForCommand(gdfDir, to, "gdf app move --to")
		if err != nil {
			return err
		}
		to = resolvedTo
	}

	if from == to {
		return fmt.Errorf("source and target profiles are the same ('%s')", from)
	}

	// 2. Load Profiles
	sourcePath := filepath.Join(gdfDir, "profiles", from, "profile.yaml")
	sourceProfile, err := config.LoadProfile(sourcePath)
	if err != nil {
		return fmt.Errorf("loading source profile '%s': %w", from, err)
	}

	// Target might not exist
	targetDir := filepath.Join(gdfDir, "profiles", to)
	targetPath := filepath.Join(targetDir, "profile.yaml")
	targetProfile, err := config.LoadProfile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			targetProfile = &config.Profile{
				Name: to,
				Apps: []string{},
			}
			// Ensure directory exists
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("creating target profile directory: %w", err)
			}
		} else {
			return fmt.Errorf("loading target profile '%s': %w", to, err)
		}
	}

	// 3. Find matches
	var matchedApps []string
	var remainingApps []string

	for _, app := range sourceProfile.Apps {
		matched, err := filepath.Match(pattern, app)
		if err != nil {
			return fmt.Errorf("invalid pattern '%s': %w", pattern, err)
		}
		if matched {
			matchedApps = append(matchedApps, app)
		} else {
			remainingApps = append(remainingApps, app)
		}
	}

	if len(matchedApps) == 0 {
		fmt.Printf("No apps matched '%s' in profile '%s'\n", pattern, from)
		return nil
	}

	// 4. Move Apps
	// Add to target (deduplicate)
	existingTargetApps := make(map[string]bool)
	for _, app := range targetProfile.Apps {
		existingTargetApps[app] = true
	}

	movedCount := 0
	for _, app := range matchedApps {
		if !existingTargetApps[app] {
			targetProfile.Apps = append(targetProfile.Apps, app)
			movedCount++
		} else {
			fmt.Printf("App '%s' already exists in target profile '%s', skipping add (removing from source)\n", app, to)
		}
	}

	// Remove from source
	sourceProfile.Apps = remainingApps

	// 5. Save
	if err := sourceProfile.Save(sourcePath); err != nil {
		return fmt.Errorf("saving source profile: %w", err)
	}
	if err := targetProfile.Save(targetPath); err != nil {
		return fmt.Errorf("saving target profile: %w", err)
	}

	fmt.Printf("✓ Moved %d apps from '%s' to '%s'\n", len(matchedApps), from, to)
	for _, app := range matchedApps {
		fmt.Printf("  - %s\n", app)
	}
	if moveApply {
		return applyProfilesGuarded([]string{from, to})
	}

	return nil
}
