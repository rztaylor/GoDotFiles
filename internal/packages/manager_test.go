package packages

import (
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/platform"
)

func TestForPlatform(t *testing.T) {
	tests := []struct {
		name        string
		platform    *platform.Platform
		wantManager string
	}{
		{
			name: "macOS returns brew",
			platform: &platform.Platform{
				OS: "macos",
			},
			wantManager: "brew",
		},
		{
			name: "Ubuntu returns apt",
			platform: &platform.Platform{
				OS:     "linux",
				Distro: "ubuntu",
			},
			wantManager: "apt",
		},
		{
			name: "Debian returns apt",
			platform: &platform.Platform{
				OS:     "linux",
				Distro: "debian",
			},
			wantManager: "apt",
		},
		{
			name: "Fedora returns dnf",
			platform: &platform.Platform{
				OS:     "linux",
				Distro: "fedora",
			},
			wantManager: "none", // dnf not available on this system
		},
		{
			name: "RHEL returns dnf",
			platform: &platform.Platform{
				OS:     "linux",
				Distro: "rhel",
			},
			wantManager: "none", // dnf not available on this system
		},
		{
			name: "WSL with Ubuntu returns apt",
			platform: &platform.Platform{
				OS:     "wsl",
				Distro: "ubuntu",
			},
			wantManager: "apt",
		},
		{
			name: "unknown platform returns noop",
			platform: &platform.Platform{
				OS:     "unknown",
				Distro: "",
			},
			wantManager: "none",
		},
		{
			name: "Arch Linux returns noop (not yet implemented)",
			platform: &platform.Platform{
				OS:     "linux",
				Distro: "arch",
			},
			wantManager: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := ForPlatform(tt.platform)
			if mgr.Name() != tt.wantManager {
				t.Errorf("ForPlatform() manager = %q, want %q", mgr.Name(), tt.wantManager)
			}
		})
	}
}

func TestNoOpManager(t *testing.T) {
	mgr := &NoOpManager{}

	// Test Name
	if got := mgr.Name(); got != "none" {
		t.Errorf("NoOpManager.Name() = %q, want %q", got, "none")
	}

	// Test Install - should not error
	if err := mgr.Install("anything"); err != nil {
		t.Errorf("NoOpManager.Install() error = %v, want nil", err)
	}

	// Test IsInstalled - should always return false
	installed, err := mgr.IsInstalled("anything")
	if err != nil {
		t.Errorf("NoOpManager.IsInstalled() error = %v, want nil", err)
	}
	if installed {
		t.Error("NoOpManager.IsInstalled() = true, want false")
	}
}
