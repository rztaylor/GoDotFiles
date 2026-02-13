package packages

import (
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

func TestInstaller_SelectManager(t *testing.T) {
	tests := []struct {
		name        string
		pkg         *apps.Package
		platform    *platform.Platform
		wantManager string // Manager name
		wantErr     bool
	}{
		{
			name: "macOS uses brew",
			pkg: &apps.Package{
				Brew: "kubectl",
			},
			platform: &platform.Platform{
				OS: "macos",
			},
			wantManager: "brew",
			wantErr:     false,
		},
		{
			name: "Ubuntu uses apt",
			pkg: &apps.Package{
				Apt: &apps.AptPackage{Name: "kubectl"},
			},
			platform: &platform.Platform{
				OS:     "linux",
				Distro: "ubuntu",
			},
			wantManager: "apt",
			wantErr:     false,
		},
		{
			name: "Fedora uses dnf",
			pkg: &apps.Package{
				Dnf: "kubectl",
			},
			platform: &platform.Platform{
				OS:     "linux",
				Distro: "fedora",
			},
			wantManager: "dnf",
			wantErr:     false,
		},
		{
			name: "prefer override - use brew on linux",
			pkg: &apps.Package{
				Brew: "kubectl",
				Apt:  &apps.AptPackage{Name: "kubectl"},
				Prefer: &apps.Prefer{
					Linux: "brew",
				},
			},
			platform: &platform.Platform{
				OS:     "linux",
				Distro: "ubuntu",
			},
			wantManager: "brew",
			wantErr:     false,
		},
		{
			name: "WSL uses apt by default",
			pkg: &apps.Package{
				Apt: &apps.AptPackage{Name: "tree"},
			},
			platform: &platform.Platform{
				OS:     "wsl",
				Distro: "ubuntu",
			},
			wantManager: "apt",
			wantErr:     false,
		},
		{
			name: "no matching package manager",
			pkg: &apps.Package{
				Brew: "only-on-mac",
			},
			platform: &platform.Platform{
				OS:     "linux",
				Distro: "ubuntu",
			},
			wantManager: "none",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installer := NewInstaller()
			mgr := installer.SelectManager(tt.pkg, tt.platform)

			if mgr.Name() != tt.wantManager {
				t.Errorf("SelectManager() manager = %q, want %q", mgr.Name(), tt.wantManager)
			}
		})
	}
}

func TestInstaller_Install_Validation(t *testing.T) {
	installer := NewInstaller()

	// Test nil package
	err := installer.Install(nil, &platform.Platform{OS: "macos"})
	if err == nil {
		t.Error("Install(nil, ...) should return error")
	}
}

func TestInstaller_IsInstalled_Validation(t *testing.T) {
	installer := NewInstaller()

	// Test nil package
	_, err := installer.IsInstalled(nil, &platform.Platform{OS: "macos"})
	if err == nil {
		t.Error("IsInstalled(nil, ...) should return error")
	}
}
