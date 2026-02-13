// Package platform provides OS detection and platform abstraction.
//
// # Responsibility
//
// This package handles:
//   - Detecting the current OS (macOS, Linux, WSL)
//   - Detecting Linux distributions (Ubuntu, Fedora, Arch)
//   - Path normalization (expanding ~, XDG directories)
//   - Hostname and architecture detection
//
// # Key Types
//
//   - Platform: Contains OS, distro, hostname, arch info
//   - PathExpander: Handles path expansion and normalization
//
// # Usage
//
//	p := platform.Detect()
//	if p.OS == "wsl" {
//	    // WSL-specific handling
//	}
package platform
