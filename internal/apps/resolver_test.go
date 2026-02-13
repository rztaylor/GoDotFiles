package apps

import (
	"testing"
)

func TestResolveApps(t *testing.T) {
	tests := []struct {
		name        string
		apps        map[string]*Bundle
		requested   []string
		want        []string // Expected order
		wantErr     bool
		errContains string
	}{
		{
			name: "single app no dependencies",
			apps: map[string]*Bundle{
				"git": {Name: "git"},
			},
			requested: []string{"git"},
			want:      []string{"git"},
			wantErr:   false,
		},
		{
			name: "linear dependencies",
			apps: map[string]*Bundle{
				"base":   {Name: "base"},
				"tools":  {Name: "tools", Dependencies: []string{"base"}},
				"custom": {Name: "custom", Dependencies: []string{"tools"}},
			},
			requested: []string{"custom"},
			want:      []string{"base", "tools", "custom"},
			wantErr:   false,
		},
		{
			name: "multiple apps with shared dependency",
			apps: map[string]*Bundle{
				"base": {Name: "base"},
				"vim":  {Name: "vim", Dependencies: []string{"base"}},
				"git":  {Name: "git", Dependencies: []string{"base"}},
			},
			requested: []string{"vim", "git"},
			want:      []string{"base", "vim", "git"},
			wantErr:   false,
		},
		{
			name: "diamond dependency",
			apps: map[string]*Bundle{
				"base":  {Name: "base"},
				"left":  {Name: "left", Dependencies: []string{"base"}},
				"right": {Name: "right", Dependencies: []string{"base"}},
				"top":   {Name: "top", Dependencies: []string{"left", "right"}},
			},
			requested: []string{"top"},
			want:      []string{"base", "left", "right", "top"},
			wantErr:   false,
		},
		{
			name: "circular dependency",
			apps: map[string]*Bundle{
				"a": {Name: "a", Dependencies: []string{"b"}},
				"b": {Name: "b", Dependencies: []string{"a"}},
			},
			requested:   []string{"a"},
			wantErr:     true,
			errContains: "circular dependency",
		},
		{
			name: "self reference",
			apps: map[string]*Bundle{
				"self": {Name: "self", Dependencies: []string{"self"}},
			},
			requested:   []string{"self"},
			wantErr:     true,
			errContains: "circular dependency",
		},
		{
			name: "missing app",
			apps: map[string]*Bundle{
				"git": {Name: "git"},
			},
			requested:   []string{"missing"},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "missing dependency",
			apps: map[string]*Bundle{
				"app": {Name: "app", Dependencies: []string{"missing"}},
			},
			requested:   []string{"app"},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveApps(tt.requested, tt.apps)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveApps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ResolveApps() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("ResolveApps() got %d apps, want %d", len(got), len(tt.want))
				return
			}

			for i, bundle := range got {
				if bundle.Name != tt.want[i] {
					t.Errorf("ResolveApps()[%d] = %s, want %s", i, bundle.Name, tt.want[i])
				}
			}
		})
	}
}

func TestBundleMap(t *testing.T) {
	bundles := []*Bundle{
		{Name: "git"},
		{Name: "vim"},
		{Name: "kubectl"},
	}

	m := BundleMap(bundles)

	if len(m) != 3 {
		t.Errorf("BundleMap() got %d entries, want 3", len(m))
	}

	for _, b := range bundles {
		if m[b.Name] != b {
			t.Errorf("BundleMap()[%s] not found or incorrect", b.Name)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
