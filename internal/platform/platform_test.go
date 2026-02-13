package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	p := Detect()

	// Should always have values
	if p.OS == "" {
		t.Error("OS should not be empty")
	}
	if p.Arch == "" {
		t.Error("Arch should not be empty")
	}
	if p.Home == "" && os.Getenv("HOME") != "" {
		t.Error("Home should not be empty when HOME is set")
	}

	// OS should be one of the expected values
	validOS := map[string]bool{"macos": true, "linux": true, "wsl": true}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !validOS[p.OS] {
			t.Errorf("OS = %q, want one of macos/linux/wsl", p.OS)
		}
	}
}

func TestParseOSRelease(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "ubuntu",
			content: `NAME="Ubuntu"
VERSION="22.04.1 LTS (Jammy Jellyfish)"
ID=ubuntu
`,
			want: "ubuntu",
		},
		{
			name: "fedora",
			content: `NAME="Fedora Linux"
VERSION="38 (Workstation Edition)"
ID=fedora
`,
			want: "fedora",
		},
		{
			name: "arch",
			content: `NAME="Arch Linux"
ID=arch
`,
			want: "arch",
		},
		{
			name: "quoted id",
			content: `NAME="Some Distro"
ID="custom"
`,
			want: "custom",
		},
		{
			name:    "empty",
			content: "",
			want:    "",
		},
		{
			name:    "no id",
			content: "NAME=Test",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOSRelease(tt.content)
			if got != tt.want {
				t.Errorf("parseOSRelease() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlatform_Is(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		isMacOS  bool
		isLinux  bool
		isWSL    bool
		isDebian bool
		isFedora bool
		isArch   bool
	}{
		{
			name:     "macos",
			platform: Platform{OS: "macos"},
			isMacOS:  true,
		},
		{
			name:     "linux",
			platform: Platform{OS: "linux"},
			isLinux:  true,
		},
		{
			name:     "wsl",
			platform: Platform{OS: "wsl"},
			isWSL:    true,
		},
		{
			name:     "ubuntu",
			platform: Platform{OS: "linux", Distro: "ubuntu"},
			isLinux:  true,
			isDebian: true,
		},
		{
			name:     "fedora",
			platform: Platform{OS: "linux", Distro: "fedora"},
			isLinux:  true,
			isFedora: true,
		},
		{
			name:     "arch",
			platform: Platform{OS: "linux", Distro: "arch"},
			isLinux:  true,
			isArch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.IsMacOS(); got != tt.isMacOS {
				t.Errorf("IsMacOS() = %v, want %v", got, tt.isMacOS)
			}
			if got := tt.platform.IsLinux(); got != tt.isLinux {
				t.Errorf("IsLinux() = %v, want %v", got, tt.isLinux)
			}
			if got := tt.platform.IsWSL(); got != tt.isWSL {
				t.Errorf("IsWSL() = %v, want %v", got, tt.isWSL)
			}
			if got := tt.platform.IsDebian(); got != tt.isDebian {
				t.Errorf("IsDebian() = %v, want %v", got, tt.isDebian)
			}
			if got := tt.platform.IsFedora(); got != tt.isFedora {
				t.Errorf("IsFedora() = %v, want %v", got, tt.isFedora)
			}
			if got := tt.platform.IsArch(); got != tt.isArch {
				t.Errorf("IsArch() = %v, want %v", got, tt.isArch)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME not set")
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "tilde expansion",
			path: "~/.config",
			want: filepath.Join(home, ".config"),
		},
		{
			name: "just tilde",
			path: "~",
			want: home,
		},
		{
			name: "no expansion needed",
			path: "/absolute/path",
			want: "/absolute/path",
		},
		{
			name: "empty",
			path: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPath(tt.path)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestExpandPathWithHome(t *testing.T) {
	tests := []struct {
		name string
		path string
		home string
		want string
	}{
		{
			name: "tilde with custom home",
			path: "~/.config",
			home: "/custom/home",
			want: "/custom/home/.config",
		},
		{
			name: "no tilde",
			path: "/absolute/path",
			home: "/custom/home",
			want: "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPathWithHome(tt.path, tt.home)
			if got != tt.want {
				t.Errorf("ExpandPathWithHome(%q, %q) = %q, want %q", tt.path, tt.home, got, tt.want)
			}
		})
	}
}

func TestConfigDir(t *testing.T) {
	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME not set")
	}

	want := filepath.Join(home, ".gdf")
	got := ConfigDir()
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}
