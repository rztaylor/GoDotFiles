// Package packages provides abstractions over system package managers.
//
// # Responsibility
//
// This package handles:
//   - Installing/uninstalling packages via brew, apt, dnf
//   - Checking if packages are installed
//   - Running custom install scripts with user confirmation
//   - Handling OS-specific package names
//   - Repository and GPG key management for apt
//
// # Key Types
//
//   - Manager: Interface for package managers
//   - Brew: Homebrew/Linuxbrew implementation
//   - Apt: Debian/Ubuntu apt implementation
//   - Dnf: Fedora/RHEL dnf implementation
//   - Custom: Custom script-based installation with security controls
//   - Installer: Coordinator for selecting and using package managers
//
// # Usage
//
// Using a specific package manager:
//
//	mgr := packages.ForPlatform(platform)
//	if err := mgr.Install("kubectl"); err != nil {
//	    return err
//	}
//
// Using the Installer (recommended):
//
//	installer := packages.NewInstaller()
//	pkg := &apps.Package{
//	    Brew: "kubectl",
//	    Apt: &apps.AptPackage{Name: "kubectl"},
//	}
//	if err := installer.Install(pkg, platform); err != nil {
//	    return err
//	}
//
// # Security Model
//
// Custom installation scripts require explicit user confirmation:
//   - Always prompts before execution (cannot be bypassed)
//   - Displays script source for review
//   - Clearly indicates when sudo is required
//   - Logs all executions for audit trail
package packages
