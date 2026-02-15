package packages

import (
	"os/exec"

	"github.com/rztaylor/GoDotFiles/internal/platform"
)

// lookPath allows mocking exec.LookPath in tests
var lookPath = exec.LookPath

// Manager is the interface for package managers.
type Manager interface {
	// Install installs a package.
	Install(pkg string) error

	// Uninstall removes a package.
	Uninstall(pkg string) error

	// IsInstalled checks if a package is installed.
	IsInstalled(pkg string) (bool, error)

	// Name returns the package manager name.
	Name() string
}

// NoOpManager is a package manager that does nothing.
// Used when no package manager is available or during graceful degradation.
type NoOpManager struct{}

// Install returns nil (no-op).
func (m *NoOpManager) Install(pkg string) error {
	return nil
}

// Uninstall returns nil (no-op).
func (m *NoOpManager) Uninstall(pkg string) error {
	return nil
}

// IsInstalled always returns false (no-op).
func (m *NoOpManager) IsInstalled(pkg string) (bool, error) {
	return false, nil
}

// Name returns "none".
func (m *NoOpManager) Name() string {
	return "none"
}

// Override allows tests to force a specific package manager.
var Override Manager

// ForPlatform returns an appropriate package manager for the platform.
func ForPlatform(p *platform.Platform) Manager {
	if Override != nil {
		return Override
	}
	switch {
	case p.IsMacOS():
		brew := NewBrew()
		if brew.IsAvailable() {
			return brew
		}
	case p.IsDebian():
		apt := NewApt()
		if apt.IsAvailable() {
			return apt
		}
	case p.IsFedora():
		dnf := NewDnf()
		if dnf.IsAvailable() {
			return dnf
		}
	}

	// Fall back to NoOpManager if no package manager available
	return &NoOpManager{}
}
