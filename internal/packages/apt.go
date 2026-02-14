package packages

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

// Apt implements the Manager interface for Debian/Ubuntu apt.
type Apt struct {
	// execCommand allows mocking in tests
	execCommand func(string, ...string) *exec.Cmd
}

// NewApt creates a new Apt package manager.
func NewApt() *Apt {
	return &Apt{
		execCommand: exec.Command,
	}
}

// Install installs a package using apt-get.
func (a *Apt) Install(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	execCmd := a.execCommand
	if execCmd == nil {
		execCmd = exec.Command
	}

	// Use sudo apt-get install -y
	cmd := execCmd("sudo", "apt-get", "install", "-y", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s via apt: %w\nOutput: %s", pkg, err, string(output))
	}

	return nil
}

// InstallWithRepo installs a package with optional repository and key setup.
func (a *Apt) InstallWithRepo(aptPkg *apps.AptPackage) error {
	if aptPkg == nil {
		return fmt.Errorf("apt package configuration cannot be nil")
	}

	if aptPkg.Name == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	execCmd := a.execCommand
	if execCmd == nil {
		execCmd = exec.Command
	}

	// Add GPG key if specified
	if aptPkg.Key != "" {
		cmd := execCmd("sh", "-c", fmt.Sprintf("curl -fsSL %s | sudo apt-key add -", aptPkg.Key))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to add GPG key: %w\nOutput: %s", err, string(output))
		}
	}

	// Add repository if specified
	if aptPkg.Repo != "" {
		cmd := execCmd("sudo", "add-apt-repository", "-y", aptPkg.Repo)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to add repository: %w\nOutput: %s", err, string(output))
		}

		// Update package lists after adding repo
		cmd = execCmd("sudo", "apt-get", "update")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to update package lists: %w\nOutput: %s", err, string(output))
		}
	}

	// Install the package
	return a.Install(aptPkg.Name)
}

// IsInstalled checks if a package is installed via apt.
func (a *Apt) IsInstalled(pkg string) (bool, error) {
	if pkg == "" {
		return false, fmt.Errorf("package name cannot be empty")
	}

	execCmd := a.execCommand
	if execCmd == nil {
		execCmd = exec.Command
	}

	cmd := execCmd("dpkg", "-l", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// dpkg -l returns error if package not found
		return false, nil
	}

	// Check if package is actually installed (status "ii")
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ii") && strings.Contains(line, pkg) {
			return true, nil
		}
	}

	return false, nil
}

// Name returns the package manager name.
func (a *Apt) Name() string {
	return "apt"
}

// IsAvailable checks if apt is available on the system.
func (a *Apt) IsAvailable() bool {
	_, err := lookPath("apt-get")
	return err == nil
}
