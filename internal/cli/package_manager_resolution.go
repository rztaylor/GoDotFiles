package cli

import (
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

type packageManagerCandidate struct {
	Name        string
	Manager     packages.Manager
	PackageName string
}

type packageManagerPlan struct {
	Selected packageManagerCandidate
	Probes   []packageManagerCandidate
}

var packageManagerFactory = defaultPackageManagerFactory
var packageAutoManagerForPlatform = packages.ForPlatform

func resolvePackageManagerPlan(pkg *apps.Package, plat *platform.Platform, cfg *config.Config) *packageManagerPlan {
	if pkg == nil || plat == nil {
		return nil
	}

	preferredFromApp := appPreferredManagerName(pkg.Prefer, plat)
	preferredFromConfig := configPreferredManagerName(cfg, plat)
	autoName := normalizeManagerName(packageAutoManagerForPlatform(plat).Name())

	candidates := map[string]packageManagerCandidate{}
	addCandidate := func(name string) bool {
		name = normalizeManagerName(name)
		if name == "" {
			return false
		}
		if _, ok := candidates[name]; ok {
			return true
		}
		pkgName, defined := pkg.ResolveName(name)
		if !defined || pkgName == "" {
			return false
		}
		manager, available := packageManagerFactory(name)
		if !available || manager == nil {
			return false
		}
		candidates[name] = packageManagerCandidate{
			Name:        name,
			Manager:     manager,
			PackageName: pkgName,
		}
		return true
	}

	selectedName := ""
	for _, name := range []string{preferredFromApp, preferredFromConfig, autoName} {
		if addCandidate(name) {
			selectedName = normalizeManagerName(name)
			break
		}
	}

	// Fallback to first available configured manager for this app.
	for _, name := range supportedManagersInProbeOrder() {
		addCandidate(name)
		if selectedName == "" {
			if _, ok := candidates[name]; ok {
				selectedName = name
			}
		}
	}

	if selectedName == "" {
		return nil
	}

	probes := []packageManagerCandidate{candidates[selectedName]}
	for _, name := range supportedManagersInProbeOrder() {
		if name == selectedName {
			continue
		}
		if c, ok := candidates[name]; ok {
			probes = append(probes, c)
		}
	}

	return &packageManagerPlan{
		Selected: candidates[selectedName],
		Probes:   probes,
	}
}

func defaultPackageManagerFactory(name string) (packages.Manager, bool) {
	name = normalizeManagerName(name)

	if packages.Override != nil && normalizeManagerName(packages.Override.Name()) == name {
		return packages.Override, true
	}

	switch name {
	case "brew":
		mgr := packages.NewBrew()
		return mgr, mgr.IsAvailable()
	case "apt":
		mgr := packages.NewApt()
		return mgr, mgr.IsAvailable()
	case "dnf":
		mgr := packages.NewDnf()
		return mgr, mgr.IsAvailable()
	default:
		return nil, false
	}
}

func appPreferredManagerName(prefer *apps.Prefer, plat *platform.Platform) string {
	if prefer == nil || plat == nil {
		return ""
	}
	switch {
	case plat.IsMacOS():
		return normalizeManagerName(prefer.Macos)
	case plat.IsWSL():
		if prefer.Wsl != "" {
			return normalizeManagerName(prefer.Wsl)
		}
		return normalizeManagerName(prefer.Linux)
	default:
		return normalizeManagerName(prefer.Linux)
	}
}

func configPreferredManagerName(cfg *config.Config, plat *platform.Platform) string {
	if cfg == nil || cfg.PackageManager == nil || cfg.PackageManager.Prefer == nil || plat == nil {
		return ""
	}
	prefer := cfg.PackageManager.Prefer
	switch {
	case plat.IsMacOS():
		return normalizeManagerName(prefer.Macos)
	case plat.IsWSL():
		if prefer.Wsl != "" {
			return normalizeManagerName(prefer.Wsl)
		}
		return normalizeManagerName(prefer.Linux)
	default:
		return normalizeManagerName(prefer.Linux)
	}
}

func normalizeManagerName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func supportedManagersInProbeOrder() []string {
	return []string{"brew", "apt", "dnf"}
}
