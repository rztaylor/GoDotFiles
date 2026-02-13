package packages

import (
	"fmt"
	"os/exec"
)

// Brew implements the Manager interface for Homebrew/Linuxbrew.
type Brew struct {
	// execCommand allows mocking in tests
	execCommand func(string, ...string) *exec.Cmd
}

// NewBrew creates a new Brew package manager.
func NewBrew() *Brew {
	return &Brew{
		execCommand: exec.Command,
	}
}

// Install installs a package using brew.
func (b *Brew) Install(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	execCmd := b.execCommand
	if execCmd == nil {
		execCmd = exec.Command
	}

	cmd := execCmd("brew", "install", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s via brew: %w\nOutput: %s", pkg, err, string(output))
	}

	return nil
}

// IsInstalled checks if a package is installed via brew.
func (b *Brew) IsInstalled(pkg string) (bool, error) {
	if pkg == "" {
		return false, fmt.Errorf("package name cannot be empty")
	}

	execCmd := b.execCommand
	if execCmd == nil {
		execCmd = exec.Command
	}

	cmd := execCmd("brew", "list", pkg)
	err := cmd.Run()
	if err != nil {
		// Exit code 1 means not installed
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		// Other errors are actual failures
		return false, fmt.Errorf("failed to check if %s is installed: %w", pkg, err)
	}

	return true, nil
}

// Name returns the package manager name.
func (b *Brew) Name() string {
	return "brew"
}

// IsAvailable checks if brew is available on the system.
func (b *Brew) IsAvailable() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}
