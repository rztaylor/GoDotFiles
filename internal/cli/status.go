package cli

import (
	"fmt"
	"time"

	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show applied profiles and their status",
	Long: `Display which profiles are currently applied and when they were last applied.

This command shows:
  - List of applied profiles with their apps
  - Total number of apps across all profiles
  - When profiles were last applied

The state is stored locally in ~/.gdf/state.yaml and does not sync across machines.`,
	Example: `  gdf status`,
	RunE:    runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()

	// Load state
	st, err := state.LoadFromDir(gdfDir)
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	// Check if any profiles are applied
	if len(st.AppliedProfiles) == 0 {
		fmt.Println("No profiles currently applied.")
		fmt.Println()
		fmt.Println("💡 Use 'gdf apply <profile>' to apply a profile.")
		return nil
	}

	// Display applied profiles
	fmt.Println("Applied Profiles:")
	for _, profile := range st.AppliedProfiles {
		appCount := len(profile.Apps)
		timeAgo := formatTimeAgo(profile.AppliedAt)
		fmt.Printf("  ✓ %s (%d app%s) - applied %s\n",
			profile.Name,
			appCount,
			pluralize(appCount),
			timeAgo,
		)
	}

	// Display all apps
	allApps := st.GetAppliedApps()
	fmt.Printf("\nApps (%d total):\n", len(allApps))
	fmt.Print("  ")
	for i, app := range allApps {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(app)
	}
	fmt.Println()

	// Display last applied time
	if !st.LastApplied.IsZero() {
		fmt.Printf("\nLast applied: %s\n", st.LastApplied.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// formatTimeAgo formats a time as a human-readable "time ago" string.
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, pluralize(minutes))
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, pluralize(hours))
	}
	days := int(duration.Hours() / 24)
	return fmt.Sprintf("%d day%s ago", days, pluralize(days))
}

// pluralize returns "s" if count != 1, otherwise empty string.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
