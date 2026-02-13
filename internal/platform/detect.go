package platform

import (
	"os"
	"runtime"
	"strings"
)

// Platform contains information about the current platform.
type Platform struct {
	// OS is the operating system: macos, linux, or wsl
	OS string

	// Distro is the Linux distribution (empty on macOS): ubuntu, fedora, arch, etc.
	Distro string

	// Hostname is the machine hostname.
	Hostname string

	// Arch is the CPU architecture: amd64, arm64, etc.
	Arch string

	// Home is the user's home directory.
	Home string
}

// Override allows tests to force a specific platform.
var Override *Platform

// Detect returns information about the current platform.
func Detect() *Platform {
	if Override != nil {
		return Override
	}

	p := &Platform{
		Arch: runtime.GOARCH,
		Home: os.Getenv("HOME"),
	}

	hostname, _ := os.Hostname()
	p.Hostname = hostname

	// Detect OS
	switch runtime.GOOS {
	case "darwin":
		p.OS = "macos"
	case "linux":
		if isWSL() {
			p.OS = "wsl"
		} else {
			p.OS = "linux"
		}
		p.Distro = detectDistro()
	default:
		p.OS = runtime.GOOS
	}

	return p
}

// isWSL checks if running inside Windows Subsystem for Linux.
func isWSL() bool {
	// Check /proc/version for Microsoft/WSL indicators
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	version := strings.ToLower(string(data))
	return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
}

// detectDistro detects the Linux distribution.
func detectDistro() string {
	// Try /etc/os-release first (most modern distros)
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		return parseOSRelease(string(data))
	}

	// Fallback: check for specific files
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return "debian"
	}
	if _, err := os.Stat("/etc/fedora-release"); err == nil {
		return "fedora"
	}
	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return "arch"
	}

	return ""
}

// DetectShell returns the user's current shell (bash, zsh, or unknown).
func DetectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "unknown"
	}

	// Extract shell name from path (e.g., /bin/bash -> bash)
	parts := strings.Split(shell, "/")
	if len(parts) > 0 {
		shellName := parts[len(parts)-1]
		// Normalize common shell names
		switch shellName {
		case "bash", "zsh", "fish":
			return shellName
		}
	}

	return "unknown"
}

// parseOSRelease parses /etc/os-release and returns the distro ID.
func parseOSRelease(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "ID=") {
			id := strings.TrimPrefix(line, "ID=")
			id = strings.Trim(id, `"'`)
			return strings.ToLower(id)
		}
	}
	return ""
}

// IsMacOS returns true if running on macOS.
func (p *Platform) IsMacOS() bool {
	return p.OS == "macos"
}

// IsLinux returns true if running on Linux (not WSL).
func (p *Platform) IsLinux() bool {
	return p.OS == "linux"
}

// IsWSL returns true if running on Windows Subsystem for Linux.
func (p *Platform) IsWSL() bool {
	return p.OS == "wsl"
}

// IsDebian returns true if running on Debian-based distro.
func (p *Platform) IsDebian() bool {
	return p.Distro == "debian" || p.Distro == "ubuntu" || p.Distro == "linuxmint"
}

// IsFedora returns true if running on Fedora-based distro.
func (p *Platform) IsFedora() bool {
	return p.Distro == "fedora" || p.Distro == "rhel" || p.Distro == "centos"
}

// IsArch returns true if running on Arch-based distro.
func (p *Platform) IsArch() bool {
	return p.Distro == "arch" || p.Distro == "manjaro" || p.Distro == "endeavouros"
}
