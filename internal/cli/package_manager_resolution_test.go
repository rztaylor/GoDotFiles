package cli

import (
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
)

func TestResolvePackageManagerPlan(t *testing.T) {
	plat := &platform.Platform{OS: "linux", Distro: "ubuntu"}

	tests := []struct {
		name         string
		pkg          *apps.Package
		cfg          *config.Config
		auto         string
		available    map[string]bool
		wantSelected string
		wantProbes   []string
	}{
		{
			name: "app prefer overrides global prefer",
			pkg: &apps.Package{
				Brew: "git",
				Apt:  &apps.AptPackage{Name: "git"},
				Prefer: &apps.Prefer{
					Linux: "brew",
				},
			},
			cfg: &config.Config{
				PackageManager: &config.PackageManagerConfig{
					Prefer: &config.PackageManagerPrefer{Linux: "apt"},
				},
			},
			auto:         "apt",
			available:    map[string]bool{"brew": true, "apt": true},
			wantSelected: "brew",
			wantProbes:   []string{"brew", "apt"},
		},
		{
			name: "global prefer used when app prefer unset",
			pkg: &apps.Package{
				Brew: "git",
				Apt:  &apps.AptPackage{Name: "git"},
			},
			cfg: &config.Config{
				PackageManager: &config.PackageManagerConfig{
					Prefer: &config.PackageManagerPrefer{Linux: "brew"},
				},
			},
			auto:         "apt",
			available:    map[string]bool{"brew": true, "apt": true},
			wantSelected: "brew",
			wantProbes:   []string{"brew", "apt"},
		},
		{
			name: "fallback to auto when preferred manager unavailable",
			pkg: &apps.Package{
				Brew: "git",
				Apt:  &apps.AptPackage{Name: "git"},
			},
			cfg: &config.Config{
				PackageManager: &config.PackageManagerConfig{
					Prefer: &config.PackageManagerPrefer{Linux: "brew"},
				},
			},
			auto:         "apt",
			available:    map[string]bool{"brew": false, "apt": true},
			wantSelected: "apt",
			wantProbes:   []string{"apt"},
		},
		{
			name: "fallback to first available configured manager when auto has no mapping",
			pkg: &apps.Package{
				Brew: "git",
			},
			cfg:          &config.Config{},
			auto:         "apt",
			available:    map[string]bool{"brew": true, "apt": true},
			wantSelected: "brew",
			wantProbes:   []string{"brew"},
		},
		{
			name: "falls back to custom install when no package manager mapping exists",
			pkg: &apps.Package{
				Custom: &apps.CustomInstall{Script: "echo install"},
			},
			cfg:          &config.Config{},
			auto:         "apt",
			available:    map[string]bool{"brew": true, "apt": true},
			wantSelected: "custom",
			wantProbes:   []string{"custom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			managers := map[string]*MockPackageManager{
				"brew": {mgrName: "brew"},
				"apt":  {mgrName: "apt"},
				"dnf":  {mgrName: "dnf"},
			}

			oldFactory := packageManagerFactory
			oldAuto := packageAutoManagerForPlatform
			packageManagerFactory = func(name string) (packages.Manager, bool) {
				mgr := managers[name]
				if mgr == nil {
					return nil, false
				}
				return mgr, tt.available[name]
			}
			packageAutoManagerForPlatform = func(_ *platform.Platform) packages.Manager {
				return &MockPackageManager{mgrName: tt.auto}
			}
			defer func() {
				packageManagerFactory = oldFactory
				packageAutoManagerForPlatform = oldAuto
			}()

			plan := resolvePackageManagerPlan(tt.pkg, plat, tt.cfg)
			if plan == nil {
				t.Fatal("resolvePackageManagerPlan() returned nil")
			}

			if plan.Selected.Name != tt.wantSelected {
				t.Fatalf("selected manager = %q, want %q", plan.Selected.Name, tt.wantSelected)
			}

			if len(plan.Probes) != len(tt.wantProbes) {
				t.Fatalf("probe count = %d, want %d", len(plan.Probes), len(tt.wantProbes))
			}
			for i, want := range tt.wantProbes {
				if plan.Probes[i].Name != want {
					t.Fatalf("probe[%d] = %q, want %q", i, plan.Probes[i].Name, want)
				}
			}
		})
	}
}
