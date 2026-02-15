package cli

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/library"
)

type appStats struct {
	Name     string
	Dotfiles int
	Aliases  int
	Secrets  int
	Source   string
}

func collectAppStats(gdfDir string, appNames []string) []appStats {
	stats := make([]appStats, 0, len(appNames))
	for _, appName := range appNames {
		stats = append(stats, collectSingleAppStats(gdfDir, appName))
	}
	return stats
}

func collectSingleAppStats(gdfDir, appName string) appStats {
	stat := appStats{
		Name:   appName,
		Source: "missing",
	}

	bundle, source := resolveAppForStats(gdfDir, appName)
	stat.Source = source
	if bundle == nil {
		return stat
	}

	stat.Dotfiles = len(bundle.Dotfiles)
	if bundle.Shell != nil {
		stat.Aliases = len(bundle.Shell.Aliases)
	}
	for _, dotfile := range bundle.Dotfiles {
		if dotfile.Secret {
			stat.Secrets++
		}
	}

	return stat
}

func resolveAppForStats(gdfDir, appName string) (*apps.Bundle, string) {
	localPath := filepath.Join(gdfDir, "apps", appName+".yaml")
	bundle, err := apps.Load(localPath)
	if err == nil {
		return bundle, "local"
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, "invalid"
	}

	recipe, libErr := library.New().Get(appName)
	if libErr != nil {
		return nil, "missing"
	}
	return recipe.ToBundle(), "library"
}

func aggregateAppStats(stats []appStats) (dotfiles int, aliases int, secrets int) {
	for _, stat := range stats {
		dotfiles += stat.Dotfiles
		aliases += stat.Aliases
		secrets += stat.Secrets
	}
	return dotfiles, aliases, secrets
}
