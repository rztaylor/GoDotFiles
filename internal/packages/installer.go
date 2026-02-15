package packages

import (
	"fmt"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

// Installer coordinates package installation across different package managers.
type Installer struct {
	custom        *Custom
	selectManager func(pkg *apps.Package, p *platform.Platform) Manager
}

// NewInstaller creates a new package installer.
func NewInstaller() *Installer {
	return &Installer{
		custom: NewCustom(),
	}
}

func (i *Installer) managerFor(pkg *apps.Package, p *platform.Platform) Manager {
	if i.selectManager != nil {
		return i.selectManager(pkg, p)
	}
	return i.SelectManager(pkg, p)
}

// SelectManager returns the appropriate package manager for the given package and platform.
func (i *Installer) SelectManager(pkg *apps.Package, p *platform.Platform) Manager {
	if pkg == nil {
		return &NoOpManager{}
	}

	// Check for explicit preference first
	if pkg.Prefer != nil {
		preferredMgr := i.getPreferredManager(pkg, p)
		if preferredMgr != nil {
			return preferredMgr
		}
	}

	// Platform-based selection
	switch {
	case p.IsMacOS():
		if pkg.Brew != "" {
			return NewBrew()
		}
	case p.IsDebian() || (p.IsWSL() && p.IsDebian()):
		if pkg.Apt != nil {
			return NewApt()
		}
	case p.IsFedora():
		if pkg.Dnf != "" {
			return NewDnf()
		}
	}

	// No matching package manager
	return &NoOpManager{}
}

// getPreferredManager returns the preferred package manager based on Prefer settings.
func (i *Installer) getPreferredManager(pkg *apps.Package, p *platform.Platform) Manager {
	var preferred string

	switch {
	case p.IsMacOS():
		preferred = pkg.Prefer.Macos
	case p.IsWSL():
		if pkg.Prefer.Wsl != "" {
			preferred = pkg.Prefer.Wsl
		} else {
			preferred = pkg.Prefer.Linux
		}
	default:
		preferred = pkg.Prefer.Linux
	}

	switch preferred {
	case "brew":
		if pkg.Brew != "" {
			return NewBrew()
		}
	case "apt":
		if pkg.Apt != nil {
			return NewApt()
		}
	case "dnf":
		if pkg.Dnf != "" {
			return NewDnf()
		}
	}

	return nil
}

// Install installs a package using the appropriate package manager.
func (i *Installer) Install(pkg *apps.Package, p *platform.Platform) error {
	if pkg == nil {
		return fmt.Errorf("package cannot be nil")
	}

	// Try standard package manager first
	mgr := i.managerFor(pkg, p)

	if mgr.Name() != "none" {
		// Check if already installed
		if pkgName := i.getPackageName(pkg, mgr.Name()); pkgName != "" {
			installed, err := mgr.IsInstalled(pkgName)
			if err == nil && installed {
				return nil
			}
		}
		// For apt with repo configuration, use special method
		if mgr.Name() == "apt" && pkg.Apt != nil && (pkg.Apt.Repo != "" || pkg.Apt.Key != "") {
			aptMgr := mgr.(*Apt)
			return aptMgr.InstallWithRepo(pkg.Apt)
		}

		// Get package name for the selected manager
		pkgName := i.getPackageName(pkg, mgr.Name())
		if pkgName != "" {
			return mgr.Install(pkgName)
		}
	}

	// Fall back to custom script if available
	if pkg.Custom != nil {
		return i.custom.Execute(pkg.Custom)
	}

	// No installation method available
	return fmt.Errorf("no installation method available for this platform")
}

// IsInstalled checks if a package is installed.
func (i *Installer) IsInstalled(pkg *apps.Package, p *platform.Platform) (bool, error) {
	if pkg == nil {
		return false, fmt.Errorf("package cannot be nil")
	}

	mgr := i.managerFor(pkg, p)
	if mgr.Name() == "none" {
		// Can't check if custom script installed something
		return false, nil
	}

	pkgName := i.getPackageName(pkg, mgr.Name())
	if pkgName == "" {
		return false, nil
	}

	return mgr.IsInstalled(pkgName)
}

// getPackageName returns the package name for the specified manager.
func (i *Installer) getPackageName(pkg *apps.Package, mgrName string) string {
	switch mgrName {
	case "brew":
		return pkg.Brew
	case "apt":
		if pkg.Apt != nil {
			return pkg.Apt.Name
		}
	case "dnf":
		return pkg.Dnf
	}
	return ""
}
