package apps

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDotfileUnmarshalTargetVariants(t *testing.T) {
	tests := []struct {
		name       string
		yamlData   string
		wantTarget string
		wantMap    *TargetMap
	}{
		{
			name: "target as string",
			yamlData: `
dotfiles:
  - source: git/config
    target: ~/.gitconfig
`,
			wantTarget: "~/.gitconfig",
			wantMap:    nil,
		},
		{
			name: "target as platform map",
			yamlData: `
dotfiles:
  - source: azure/config
    target:
      default: ~/.azure/config
      wsl: /mnt/c/Users/user/.azure/config
`,
			wantTarget: "",
			wantMap: &TargetMap{
				Default: "~/.azure/config",
				Wsl:     "/mnt/c/Users/user/.azure/config",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload struct {
				Dotfiles []Dotfile `yaml:"dotfiles"`
			}
			if err := yaml.Unmarshal([]byte(tt.yamlData), &payload); err != nil {
				t.Fatalf("yaml.Unmarshal() error = %v", err)
			}
			if len(payload.Dotfiles) != 1 {
				t.Fatalf("dotfiles len = %d, want 1", len(payload.Dotfiles))
			}

			got := payload.Dotfiles[0]
			if got.Target != tt.wantTarget {
				t.Errorf("Target = %q, want %q", got.Target, tt.wantTarget)
			}

			if tt.wantMap == nil {
				if got.TargetMap != nil {
					t.Errorf("TargetMap = %#v, want nil", got.TargetMap)
				}
				return
			}

			if got.TargetMap == nil {
				t.Fatal("TargetMap is nil, want non-nil")
			}
			if *got.TargetMap != *tt.wantMap {
				t.Errorf("TargetMap = %#v, want %#v", got.TargetMap, tt.wantMap)
			}
		})
	}
}

func TestDotfileEffectiveTarget(t *testing.T) {
	d := Dotfile{
		TargetMap: &TargetMap{
			Default: "~/.config/app",
			Macos:   "~/Library/Application Support/app/config",
		},
	}

	if got := d.EffectiveTarget("macos"); got != "~/Library/Application Support/app/config" {
		t.Errorf("EffectiveTarget(macos) = %q, want %q", got, "~/Library/Application Support/app/config")
	}
	if got := d.EffectiveTarget("linux"); got != "~/.config/app" {
		t.Errorf("EffectiveTarget(linux) = %q, want %q", got, "~/.config/app")
	}
}
