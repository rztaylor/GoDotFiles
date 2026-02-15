package packages

import (
	"fmt"
	"os/exec"
	"strings"
)

// Dnf implements the Manager interface for Fedora/RHEL dnf.
type Dnf struct {
	// execCommand allows mocking in tests
	execCommand func(string, ...string) *exec.Cmd
}

// NewDnf creates a new Dnf package manager.
func NewDnf() *Dnf {
	return &Dnf{
		execCommand: exec.Command,
	}
}

// Install installs a package using dnf.
func (d *Dnf) Install(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	execCmd := d.execCommand
	if execCmd == nil {
		execCmd = exec.Command
	}

	// Use sudo dnf install -y
	cmd := execCmd("sudo", "dnf", "install", "-y", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s via dnf: %w\nOutput: %s", pkg, err, string(output))
	}

	return nil
}

// Uninstall removes a package using dnf.
func (d *Dnf) Uninstall(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	execCmd := d.execCommand
	if execCmd == nil {
		execCmd = exec.Command
	}

	cmd := execCmd("sudo", "dnf", "remove", "-y", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall %s via dnf: %w\nOutput: %s", pkg, err, string(output))
	}

	return nil
}

// IsInstalled checks if a package is installed via dnf.
func (d *Dnf) IsInstalled(pkg string) (bool, error) {
	if pkg == "" {
		return false, fmt.Errorf("package name cannot be empty")
	}

	execCmd := d.execCommand
	if execCmd == nil {
		execCmd = exec.Command
	}

	cmd := execCmd("dnf", "list", "installed", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Exit code 1 means not installed
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if %s is installed: %w", pkg, err)
	}

	// Check output for installed package
	outputStr := string(output)
	return strings.Contains(outputStr, pkg), nil
}

// Name returns the package manager name.
func (d *Dnf) Name() string {
	return "dnf"
}

// IsAvailable checks if dnf is available on the system.
func (d *Dnf) IsAvailable() bool {
	_, err := lookPath("dnf")
	return err == nil
}
