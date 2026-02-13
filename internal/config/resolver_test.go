package config

import (
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/platform"
)

func TestResolveProfiles(t *testing.T) {
	defaultPlatform := &platform.Platform{OS: "linux", Distro: "ubuntu", Arch: "amd64"}

	tests := []struct {
		name     string
		profiles map[string]*Profile
		platform *platform.Platform
		request  []string
		want     []string
		wantApps []string // Optional: check apps if relevant
		wantErr  bool
	}{
		{
			name: "single profile",
			profiles: map[string]*Profile{
				"base": {Name: "base"},
			},
			request: []string{"base"},
			want:    []string{"base"},
		},
		{
			name: "profile with includes",
			profiles: map[string]*Profile{
				"base": {Name: "base"},
				"sre":  {Name: "sre", Includes: []string{"base"}},
			},
			request: []string{"sre"},
			want:    []string{"base", "sre"},
		},
		{
			name: "conditional include (match)",
			profiles: map[string]*Profile{
				"base":        {Name: "base"},
				"linux-stuff": {Name: "linux-stuff"},
				"dev": {
					Name:     "dev",
					Includes: []string{"base"},
					Conditions: []ProfileCondition{
						{If: "os == linux", Includes: []string{"linux-stuff"}},
					},
				},
			},
			platform: &platform.Platform{OS: "linux"},
			request:  []string{"dev"},
			want:     []string{"base", "linux-stuff", "dev"},
		},
		{
			name: "conditional include (no match)",
			profiles: map[string]*Profile{
				"base":        {Name: "base"},
				"linux-stuff": {Name: "linux-stuff"},
				"dev": {
					Name:     "dev",
					Includes: []string{"base"},
					Conditions: []ProfileCondition{
						{If: "os == linux", Includes: []string{"linux-stuff"}},
					},
				},
			},
			platform: &platform.Platform{OS: "macos"},
			request:  []string{"dev"},
			want:     []string{"base", "dev"},
		},
		{
			name: "conditional app add (match)",
			profiles: map[string]*Profile{
				"dev": {
					Name: "dev",
					Apps: []string{"vim"},
					Conditions: []ProfileCondition{
						{If: "os == linux", IncludeApps: []string{"htop"}},
					},
				},
			},
			platform: &platform.Platform{OS: "linux"},
			request:  []string{"dev"},
			want:     []string{"dev"},
			wantApps: []string{"vim", "htop"},
		},
		{
			name: "conditional app remove (match)",
			profiles: map[string]*Profile{
				"dev": {
					Name: "dev",
					Apps: []string{"vim", "notepad"},
					Conditions: []ProfileCondition{
						{If: "os == linux", ExcludeApps: []string{"notepad"}},
					},
				},
			},
			platform: &platform.Platform{OS: "linux"},
			request:  []string{"dev"},
			want:     []string{"dev"},
			wantApps: []string{"vim"},
		},
		{
			name: "multiple profiles with shared dependency",
			profiles: map[string]*Profile{
				"base": {Name: "base"},
				"dev":  {Name: "dev", Includes: []string{"base"}},
				"sre":  {Name: "sre", Includes: []string{"base"}},
			},
			request: []string{"dev", "sre"},
			want:    []string{"base", "dev", "sre"},
		},
		{
			name: "nested includes",
			profiles: map[string]*Profile{
				"base": {Name: "base"},
				"dev":  {Name: "dev", Includes: []string{"base"}},
				"work": {Name: "work", Includes: []string{"dev"}},
			},
			request: []string{"work"},
			want:    []string{"base", "dev", "work"},
		},
		{
			name: "circular dependency",
			profiles: map[string]*Profile{
				"a": {Name: "a", Includes: []string{"b"}},
				"b": {Name: "b", Includes: []string{"a"}},
			},
			request: []string{"a"},
			wantErr: true,
		},
		{
			name: "self reference",
			profiles: map[string]*Profile{
				"self": {Name: "self", Includes: []string{"self"}},
			},
			request: []string{"self"},
			wantErr: true,
		},
		{
			name: "missing profile",
			profiles: map[string]*Profile{
				"base": {Name: "base"},
			},
			request: []string{"missing"},
			wantErr: true,
		},
		{
			name: "missing include",
			profiles: map[string]*Profile{
				"dev": {Name: "dev", Includes: []string{"missing"}},
			},
			request: []string{"dev"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plat := tt.platform
			if plat == nil {
				plat = defaultPlatform
			}

			result, err := ResolveProfiles(tt.request, tt.profiles, plat)
			if tt.wantErr {
				if err == nil {
					t.Error("ResolveProfiles() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveProfiles() error = %v", err)
			}

			if len(result) != len(tt.want) {
				t.Errorf("got %d profiles, want %d", len(result), len(tt.want))
				return
			}

			for i, p := range result {
				if p.Name != tt.want[i] {
					t.Errorf("result[%d] = %q, want %q", i, p.Name, tt.want[i])
				}
			}

			// Verify apps if specified (only checks the LAST profile in want list for simplicity in this general test)
			if tt.wantApps != nil && len(result) > 0 {
				lastProfile := result[len(result)-1]
				if len(lastProfile.Apps) != len(tt.wantApps) {
					t.Errorf("Apps mismatch: got %v, want %v", lastProfile.Apps, tt.wantApps)
				} else {
					for i, app := range lastProfile.Apps {
						if app != tt.wantApps[i] {
							t.Errorf("Apps[%d] = %q, want %q", i, app, tt.wantApps[i])
						}
					}
				}
			}
		})
	}
}

func TestProfileMap(t *testing.T) {
	profiles := []*Profile{
		{Name: "base"},
		{Name: "dev"},
	}

	m := ProfileMap(profiles)

	if len(m) != 2 {
		t.Errorf("ProfileMap() len = %d, want 2", len(m))
	}
	if m["base"] == nil || m["base"].Name != "base" {
		t.Error("ProfileMap() missing 'base'")
	}
	if m["dev"] == nil || m["dev"].Name != "dev" {
		t.Error("ProfileMap() missing 'dev'")
	}
}
