package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/git"
	"github.com/rztaylor/GoDotFiles/internal/library"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show applied profiles, health hints, and drift summary",
	Long: `Display which profiles are currently applied and when they were last applied.

This command shows:
  - List of applied profiles with their apps
  - Total number of apps across all profiles
  - Drift summary for managed dotfile targets
  - When profiles were last applied

The state is stored locally in ~/.gdf/state.yaml and does not sync across machines.`,
	Example: `  gdf status
  gdf status --json`,
	RunE: runStatus,
}

var statusDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show per-file drift details for managed targets",
	Example: `  gdf status diff
  gdf status diff --patch
  gdf status diff --json`,
	RunE: runStatusDiff,
}

var statusJSON bool
var statusDiffJSON bool
var statusDiffPatch bool
var statusDiffMaxBytes int64
var statusDiffMaxFiles int

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.AddCommand(statusDiffCmd)
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output status as JSON")
	statusDiffCmd.Flags().BoolVar(&statusDiffJSON, "json", false, "Output drift details as JSON")
	statusDiffCmd.Flags().BoolVar(&statusDiffPatch, "patch", false, "Include unified patch output for non-symlink drift targets")
	statusDiffCmd.Flags().Int64Var(&statusDiffMaxBytes, "max-bytes", 1024*1024, "Maximum source/target file size in bytes for patch generation")
	statusDiffCmd.Flags().IntVar(&statusDiffMaxFiles, "max-files", 20, "Maximum number of drift files to generate patches for")
}

func runStatus(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()
	if !git.IsRepository(gdfDir) {
		return fmt.Errorf("GDF repository not initialized at %s. Please run 'gdf init' first.", gdfDir)
	}
	report, err := collectStatusReport(gdfDir, driftOptions{})
	if err != nil {
		return err
	}

	if statusJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	if len(report.AppliedProfiles) == 0 {
		fmt.Println("No profiles currently applied.")
		fmt.Println()
		fmt.Println("💡 Use 'gdf apply <profile>' to apply a profile.")
		return nil
	}

	fmt.Println("Applied Profiles:")
	for _, profile := range report.AppliedProfiles {
		fmt.Printf("  ✓ %s (%d app%s) - applied %s\n",
			profile.Name,
			profile.AppCount,
			pluralize(profile.AppCount),
			profile.AppliedAgo,
		)
	}

	fmt.Printf("\nApps (%d total):\n", len(report.Apps))
	fmt.Print("  ")
	for i, app := range report.Apps {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(app)
	}
	fmt.Println()

	fmt.Printf("\nDrift summary: %d issue(s)\n", report.Drift.Total)
	if report.Drift.Total > 0 {
		fmt.Printf("  source missing: %d\n", report.Drift.SourceMissing)
		fmt.Printf("  target missing: %d\n", report.Drift.TargetMissing)
		fmt.Printf("  target mismatch: %d\n", report.Drift.TargetMismatch)
		fmt.Printf("  target not symlink: %d\n", report.Drift.TargetNotSymlink)
		fmt.Printf("  target read error: %d\n", report.Drift.TargetReadError)
		fmt.Println("  💡 Run 'gdf status diff' for details.")
	}

	if report.LastApplied != "" {
		fmt.Printf("\nLast applied: %s\n", report.LastApplied)
	}

	return nil
}

func runStatusDiff(cmd *cobra.Command, args []string) error {
	opts := driftOptions{
		IncludeIssues:  true,
		IncludePreview: true,
		IncludePatch:   statusDiffPatch,
		PatchMaxBytes:  statusDiffMaxBytes,
		PatchMaxFiles:  statusDiffMaxFiles,
	}
	report, err := collectStatusReport(platform.ConfigDir(), opts)
	if err != nil {
		return err
	}

	if statusDiffJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	if len(report.Drift.Issues) == 0 {
		fmt.Println("No drift detected.")
		return nil
	}

	fmt.Printf("Drift issues (%d):\n", len(report.Drift.Issues))
	for i, issue := range report.Drift.Issues {
		fmt.Printf("  %d. [%s] app=%s target=%s\n", i+1, issue.Type, issue.App, issue.Target)
		if issue.Source != "" {
			fmt.Printf("     source: %s\n", issue.Source)
		}
		if issue.Expected != "" {
			fmt.Printf("     expected: %s\n", issue.Expected)
		}
		if issue.Actual != "" {
			fmt.Printf("     actual: %s\n", issue.Actual)
		}
		if issue.Preview != "" {
			fmt.Printf("     preview: %s\n", issue.Preview)
		}
		if issue.Patch != "" {
			fmt.Printf("     patch:\n%s\n", issue.Patch)
		}
		if issue.PatchSkippedReason != "" {
			fmt.Printf("     patch: skipped (%s)\n", issue.PatchSkippedReason)
		}
	}

	return nil
}

type driftOptions struct {
	IncludeIssues  bool
	IncludePreview bool
	IncludePatch   bool
	PatchMaxBytes  int64
	PatchMaxFiles  int
}

type statusReport struct {
	AppliedProfiles []statusProfile `json:"applied_profiles"`
	Apps            []string        `json:"apps"`
	LastApplied     string          `json:"last_applied,omitempty"`
	Drift           driftSummary    `json:"drift"`
}

type statusProfile struct {
	Name       string `json:"name"`
	AppCount   int    `json:"app_count"`
	AppliedAt  string `json:"applied_at,omitempty"`
	AppliedAgo string `json:"applied_ago,omitempty"`
}

type driftSummary struct {
	Total            int          `json:"total"`
	SourceMissing    int          `json:"source_missing"`
	TargetMissing    int          `json:"target_missing"`
	TargetMismatch   int          `json:"target_mismatch"`
	TargetNotSymlink int          `json:"target_not_symlink"`
	TargetReadError  int          `json:"target_read_error"`
	Issues           []driftIssue `json:"issues,omitempty"`
}

type driftIssue struct {
	Type               string `json:"type"`
	App                string `json:"app"`
	Source             string `json:"source,omitempty"`
	Target             string `json:"target"`
	Expected           string `json:"expected,omitempty"`
	Actual             string `json:"actual,omitempty"`
	Preview            string `json:"preview,omitempty"`
	Patch              string `json:"patch,omitempty"`
	PatchSkippedReason string `json:"patch_skipped_reason,omitempty"`
}

func collectStatusReport(gdfDir string, opts driftOptions) (*statusReport, error) {
	st, err := state.LoadFromDir(gdfDir)
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}

	report := &statusReport{
		AppliedProfiles: make([]statusProfile, 0, len(st.AppliedProfiles)),
		Apps:            []string{},
		Drift: driftSummary{
			Issues: []driftIssue{},
		},
	}

	for _, profile := range st.AppliedProfiles {
		report.AppliedProfiles = append(report.AppliedProfiles, statusProfile{
			Name:       profile.Name,
			AppCount:   len(profile.Apps),
			AppliedAt:  profile.AppliedAt.Format(time.RFC3339),
			AppliedAgo: formatTimeAgo(profile.AppliedAt),
		})
	}

	report.Apps = st.GetAppliedApps()
	sort.Strings(report.Apps)
	if !st.LastApplied.IsZero() {
		report.LastApplied = st.LastApplied.Format("2006-01-02 15:04:05")
	}

	issues, err := collectDriftIssues(gdfDir, report.Apps, opts)
	if err != nil {
		return nil, err
	}
	for _, issue := range issues {
		switch issue.Type {
		case "source_missing":
			report.Drift.SourceMissing++
		case "target_missing":
			report.Drift.TargetMissing++
		case "target_mismatch":
			report.Drift.TargetMismatch++
		case "target_not_symlink":
			report.Drift.TargetNotSymlink++
		case "target_read_error":
			report.Drift.TargetReadError++
		}
	}
	report.Drift.Total = len(issues)
	if opts.IncludeIssues {
		report.Drift.Issues = issues
	}

	return report, nil
}

func collectDriftIssues(gdfDir string, appNames []string, opts driftOptions) ([]driftIssue, error) {
	issues := make([]driftIssue, 0)
	if len(appNames) == 0 {
		return issues, nil
	}
	if opts.PatchMaxBytes <= 0 {
		opts.PatchMaxBytes = 1024 * 1024
	}
	if opts.PatchMaxFiles <= 0 {
		opts.PatchMaxFiles = 20
	}

	plat := platform.Detect()
	lib := library.New()
	cache := loadDriftCache(gdfDir)
	cacheDirty := false
	patchCount := 0
	defer func() {
		if cacheDirty {
			_ = saveDriftCache(gdfDir, cache)
		}
	}()

	for _, appName := range appNames {
		bundle, err := loadBundleForStatus(gdfDir, appName, lib)
		if err != nil {
			continue
		}
		for _, dot := range bundle.Dotfiles {
			target := dot.EffectiveTarget(plat.OS)
			if target == "" {
				continue
			}
			targetAbs := platform.ExpandPath(target)
			sourceAbs := filepath.Join(gdfDir, "dotfiles", dot.Source)

			if _, err := os.Stat(sourceAbs); err != nil {
				issues = append(issues, driftIssue{
					Type:   "source_missing",
					App:    appName,
					Source: sourceAbs,
					Target: targetAbs,
				})
				continue
			}

			targetInfo, err := os.Lstat(targetAbs)
			if os.IsNotExist(err) {
				issues = append(issues, driftIssue{
					Type:   "target_missing",
					App:    appName,
					Source: sourceAbs,
					Target: targetAbs,
				})
				continue
			}
			if err != nil {
				issues = append(issues, driftIssue{
					Type:   "target_read_error",
					App:    appName,
					Source: sourceAbs,
					Target: targetAbs,
					Actual: err.Error(),
				})
				continue
			}

			if targetInfo.Mode()&os.ModeSymlink == 0 {
				issue := driftIssue{
					Type:   "target_not_symlink",
					App:    appName,
					Source: sourceAbs,
					Target: targetAbs,
				}
				if opts.IncludePreview || opts.IncludePatch {
					preview, patch, skipped, updated := cachedDiffDetails(cache, sourceAbs, targetAbs, opts.IncludePatch, opts.PatchMaxBytes)
					if updated {
						cacheDirty = true
					}
					issue.Preview = preview
					if opts.IncludePatch {
						if patchCount >= opts.PatchMaxFiles {
							issue.PatchSkippedReason = fmt.Sprintf("hit --max-files limit (%d)", opts.PatchMaxFiles)
						} else if patch != "" {
							issue.Patch = patch
							patchCount++
						} else if skipped != "" {
							issue.PatchSkippedReason = skipped
						}
					}
				}
				issues = append(issues, issue)
				continue
			}

			linkDest, err := os.Readlink(targetAbs)
			if err != nil {
				issues = append(issues, driftIssue{
					Type:   "target_read_error",
					App:    appName,
					Source: sourceAbs,
					Target: targetAbs,
					Actual: err.Error(),
				})
				continue
			}

			actual := linkDest
			if !filepath.IsAbs(actual) {
				actual = filepath.Clean(filepath.Join(filepath.Dir(targetAbs), actual))
			}
			expected := filepath.Clean(sourceAbs)
			if filepath.Clean(actual) != expected {
				issues = append(issues, driftIssue{
					Type:     "target_mismatch",
					App:      appName,
					Source:   sourceAbs,
					Target:   targetAbs,
					Expected: expected,
					Actual:   actual,
				})
			}
		}
	}

	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Type != issues[j].Type {
			return issues[i].Type < issues[j].Type
		}
		if issues[i].App != issues[j].App {
			return issues[i].App < issues[j].App
		}
		return issues[i].Target < issues[j].Target
	})

	return issues, nil
}

func loadBundleForStatus(gdfDir, appName string, lib *library.Manager) (*apps.Bundle, error) {
	localPath := filepath.Join(gdfDir, "apps", appName+".yaml")
	bundle, err := apps.Load(localPath)
	if err == nil {
		return bundle, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	recipe, err := lib.Get(appName)
	if err != nil {
		return nil, err
	}
	return recipe.ToBundle(), nil
}

func diffPreview(sourcePath, targetPath string) string {
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return ""
	}
	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return ""
	}
	if len(sourceData) == 0 && len(targetData) == 0 {
		return ""
	}
	if strings.ContainsRune(string(sourceData), '\x00') || strings.ContainsRune(string(targetData), '\x00') {
		return "binary content differs"
	}
	if string(sourceData) == string(targetData) {
		return "content matches but target is not linked"
	}

	sourceLines := strings.Split(string(sourceData), "\n")
	targetLines := strings.Split(string(targetData), "\n")
	limit := len(sourceLines)
	if len(targetLines) < limit {
		limit = len(targetLines)
	}
	for i := 0; i < limit; i++ {
		if sourceLines[i] != targetLines[i] {
			return fmt.Sprintf("line %d differs", i+1)
		}
	}
	if len(sourceLines) != len(targetLines) {
		return "line count differs"
	}
	return "content differs"
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
