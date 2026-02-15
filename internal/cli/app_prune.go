package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

type orphanedApp struct {
	Name             string `json:"name"`
	AppPath          string `json:"app_path"`
	DotfilesPath     string `json:"dotfiles_path,omitempty"`
	HasDotfiles      bool   `json:"has_dotfiles"`
	ReferencedByApps int    `json:"referenced_by_apps"`
}

type pruneReport struct {
	Mode        string        `json:"mode"`
	DryRun      bool          `json:"dry_run"`
	ArchiveRoot string        `json:"archive_root,omitempty"`
	Orphans     []orphanedApp `json:"orphans"`
	Pruned      []string      `json:"pruned"`
}

func runPrune(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()
	report, err := pruneOrphanedApps(gdfDir, pruneDelete, pruneDryRun, pruneYes)
	if err != nil {
		return err
	}

	if pruneJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	if len(report.Orphans) == 0 {
		fmt.Println("No orphaned apps found.")
		return nil
	}

	fmt.Printf("Orphaned apps (%d):\n", len(report.Orphans))
	for _, orphan := range report.Orphans {
		fmt.Printf("  - %s (%s)\n", orphan.Name, orphan.AppPath)
	}
	if report.DryRun {
		if report.Mode == "delete" {
			fmt.Println("Dry run only. No app definitions were deleted.")
		} else {
			fmt.Println("Dry run only. No app definitions were archived.")
		}
		return nil
	}

	if report.Mode == "delete" {
		fmt.Printf("✓ Deleted %d orphaned app definition(s)\n", len(report.Pruned))
		return nil
	}
	fmt.Printf("✓ Archived %d orphaned app definition(s) to %s\n", len(report.Pruned), report.ArchiveRoot)
	return nil
}

func pruneOrphanedApps(gdfDir string, deleteMode, dryRun, yes bool) (*pruneReport, error) {
	orphans, err := findOrphanedApps(gdfDir)
	if err != nil {
		return nil, err
	}

	report := &pruneReport{
		Mode:    "archive",
		DryRun:  dryRun,
		Orphans: orphans,
		Pruned:  []string{},
	}
	if deleteMode {
		report.Mode = "delete"
	}
	if len(orphans) == 0 || dryRun {
		return report, nil
	}

	if deleteMode && !yes {
		ok, err := confirmPromptUnsafe("Permanently delete orphaned app definitions? [y/N]: ")
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, withExitCode(fmt.Errorf("prune cancelled"), exitCodeNonInteractiveStop)
		}
	}

	archiveRoot := filepath.Join(gdfDir, "archive", "apps", time.Now().UTC().Format("20060102-150405"))
	if !deleteMode {
		report.ArchiveRoot = archiveRoot
	}

	for _, orphan := range orphans {
		if deleteMode {
			if err := os.Remove(orphan.AppPath); err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("deleting app definition %s: %w", orphan.Name, err)
			}
			if orphan.HasDotfiles {
				if err := os.RemoveAll(orphan.DotfilesPath); err != nil {
					return nil, fmt.Errorf("deleting dotfiles for %s: %w", orphan.Name, err)
				}
			}
			report.Pruned = append(report.Pruned, orphan.Name)
			continue
		}

		archiveAppPath := filepath.Join(archiveRoot, "apps", orphan.Name+".yaml")
		if err := os.MkdirAll(filepath.Dir(archiveAppPath), 0755); err != nil {
			return nil, fmt.Errorf("creating archive path: %w", err)
		}
		if err := os.Rename(orphan.AppPath, archiveAppPath); err != nil {
			return nil, fmt.Errorf("archiving app definition %s: %w", orphan.Name, err)
		}

		if orphan.HasDotfiles {
			archiveDotfilesPath := filepath.Join(archiveRoot, "dotfiles", orphan.Name)
			if err := os.MkdirAll(filepath.Dir(archiveDotfilesPath), 0755); err != nil {
				return nil, fmt.Errorf("creating archive dotfiles path: %w", err)
			}
			if err := os.Rename(orphan.DotfilesPath, archiveDotfilesPath); err != nil {
				return nil, fmt.Errorf("archiving dotfiles for %s: %w", orphan.Name, err)
			}
		}
		report.Pruned = append(report.Pruned, orphan.Name)
	}
	return report, nil
}

func findOrphanedApps(gdfDir string) ([]orphanedApp, error) {
	appsDir := filepath.Join(gdfDir, "apps")
	bundles, err := apps.LoadAll(appsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []orphanedApp{}, nil
		}
		return nil, fmt.Errorf("loading app bundles: %w", err)
	}

	profiles, err := config.LoadAllProfiles(filepath.Join(gdfDir, "profiles"))
	if err != nil {
		return nil, fmt.Errorf("loading profiles: %w", err)
	}

	referenced := make(map[string]int)
	for _, profile := range profiles {
		for _, appName := range profile.Apps {
			referenced[appName]++
		}
	}

	orphans := make([]orphanedApp, 0)
	for _, bundle := range bundles {
		if referenced[bundle.Name] > 0 {
			continue
		}
		appPath := filepath.Join(appsDir, bundle.Name+".yaml")
		dotfilesPath := filepath.Join(gdfDir, "dotfiles", bundle.Name)
		_, statErr := os.Stat(dotfilesPath)
		hasDotfiles := statErr == nil
		orphans = append(orphans, orphanedApp{
			Name:             bundle.Name,
			AppPath:          appPath,
			DotfilesPath:     dotfilesPath,
			HasDotfiles:      hasDotfiles,
			ReferencedByApps: 0,
		})
	}

	sort.Slice(orphans, func(i, j int) bool {
		return orphans[i].Name < orphans[j].Name
	})
	return orphans, nil
}

func printDanglingAppCleanupGuidance(gdfDir, appName string) {
	orphans, err := findOrphanedApps(gdfDir)
	if err != nil {
		return
	}
	for _, orphan := range orphans {
		if orphan.Name != appName {
			continue
		}
		fmt.Printf("i App '%s' is no longer referenced by any profile.\n", appName)
		fmt.Printf("    Use 'gdf app prune --dry-run' to review cleanup actions.\n")
		return
	}
}
