package apps

// Package defines how to install the app's package.
// The package field is optional - omit for package-less bundles.
type Package struct {
	// Brew is the Homebrew/Linuxbrew package name (string).
	Brew string `yaml:"brew,omitempty"`

	// Apt is the Debian/Ubuntu package. Can be a simple string or AptConfig.
	Apt *AptPackage `yaml:"apt,omitempty"`

	// Dnf is the Fedora/RHEL package name.
	Dnf string `yaml:"dnf,omitempty"`

	// Pacman is the Arch Linux package name.
	Pacman string `yaml:"pacman,omitempty"`

	// Custom defines a custom installation script.
	Custom *CustomInstall `yaml:"custom,omitempty"`

	// Prefer specifies which package manager to prefer per platform.
	Prefer *Prefer `yaml:"prefer,omitempty"`
}

// AptPackage represents apt package configuration.
// Can be a simple name or include repo/key for external packages.
type AptPackage struct {
	// Name is the package name.
	Name string `yaml:"name"`

	// Repo is the APT repository URL (optional).
	Repo string `yaml:"repo,omitempty"`

	// Key is the GPG key URL for the repository (optional).
	Key string `yaml:"key,omitempty"`
}

// CustomInstall defines a custom installation script.
type CustomInstall struct {
	// Script is the shell script to run.
	Script string `yaml:"script"`

	// Sudo indicates if the script requires root privileges.
	Sudo bool `yaml:"sudo,omitempty"`

	// Confirm requires user confirmation before running (default: true).
	Confirm *bool `yaml:"confirm,omitempty"`
}

// ConfirmDefault returns the effective confirm value (default: true).
func (c *CustomInstall) ConfirmDefault() bool {
	if c.Confirm == nil {
		return true
	}
	return *c.Confirm
}

// Prefer specifies package manager preferences per platform.
type Prefer struct {
	Macos string `yaml:"macos,omitempty"`
	Linux string `yaml:"linux,omitempty"`
	Wsl   string `yaml:"wsl,omitempty"`
}

// ResolveName returns the package name for the given manager.
// Returns (name, true) if a specific package is configured.
// Returns ("", false) if no package is configured for this manager.
func (p *Package) ResolveName(manager string) (string, bool) {
	if p == nil {
		return "", false
	}
	switch manager {
	case "brew":
		if p.Brew != "" {
			return p.Brew, true
		}
	case "apt":
		if p.Apt != nil && p.Apt.Name != "" {
			return p.Apt.Name, true
		}
	case "dnf":
		if p.Dnf != "" {
			return p.Dnf, true
		}
	case "pacman":
		if p.Pacman != "" {
			return p.Pacman, true
		}
	}
	return "", false
}
