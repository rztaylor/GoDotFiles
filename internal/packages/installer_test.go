package packages

import (
	"errors"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

type mockManager struct {
	name           string
	installed      bool
	isInstalledErr error
	installErr     error
	uninstallErr   error
	installCalls   int
	uninstallCalls int
}

func (m *mockManager) Name() string {
	return m.name
}

func (m *mockManager) Install(pkg string) error {
	m.installCalls++
	return m.installErr
}

func (m *mockManager) Uninstall(pkg string) error {
	m.uninstallCalls++
	return m.uninstallErr
}

func (m *mockManager) IsInstalled(pkg string) (bool, error) {
	if m.isInstalledErr != nil {
		return false, m.isInstalledErr
	}
	return m.installed, nil
}

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

func TestInstaller_Install_SkipsAlreadyInstalled(t *testing.T) {
	mgr := &mockManager{name: "brew", installed: true}
	installer := &Installer{
		custom: NewCustom(),
		selectManager: func(pkg *apps.Package, p *platform.Platform) Manager {
			return mgr
		},
	}

	err := installer.Install(&apps.Package{Brew: "ripgrep"}, &platform.Platform{OS: "macos"})
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}
	if mgr.installCalls != 0 {
		t.Fatalf("Install() called manager install %d times, want 0", mgr.installCalls)
	}
}

func TestInstaller_Install_InstallsWhenCheckFails(t *testing.T) {
	mgr := &mockManager{name: "brew", isInstalledErr: errors.New("probe failed")}
	installer := &Installer{
		custom: NewCustom(),
		selectManager: func(pkg *apps.Package, p *platform.Platform) Manager {
			return mgr
		},
	}

	err := installer.Install(&apps.Package{Brew: "ripgrep"}, &platform.Platform{OS: "macos"})
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}
	if mgr.installCalls != 1 {
		t.Fatalf("Install() called manager install %d times, want 1", mgr.installCalls)
	}
}
