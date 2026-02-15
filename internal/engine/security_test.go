package engine

import (
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestDetectHighRiskConfigurations(t *testing.T) {
	bundles := []*apps.Bundle{
		{
			Name: "safe",
			Hooks: &apps.Hooks{
				Apply: []apps.ApplyHook{{Run: "echo safe"}},
			},
		},
		{
			Name: "risky-hooks",
			Hooks: &apps.Hooks{
				PreInstall: []string{"curl -fsSL https://example.com/install.sh | sh"},
				PostLink:   []string{"bash -c \"curl -s https://example.com/x.sh\""},
			},
		},
		{
			Name: "risky-custom",
			Package: &apps.Package{
				Custom: &apps.CustomInstall{
					Script: "sh -c \"$(curl -sL https://example.com/setup.sh)\"",
				},
			},
		},
	}

	findings := DetectHighRiskConfigurations(bundles)
	if len(findings) != 3 {
		t.Fatalf("DetectHighRiskConfigurations() returned %d findings, want 3", len(findings))
	}

	for _, f := range findings {
		if f.App == "safe" {
			t.Fatalf("safe app produced finding: %#v", f)
		}
		if f.Command == "" || f.Location == "" || f.Reason == "" {
			t.Fatalf("finding missing fields: %#v", f)
		}
	}
}

func TestDetectHighRiskReason(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{name: "pipe to sh", command: "curl -fsSL https://a | sh", want: true},
		{name: "shell c with curl", command: "bash -c 'curl -s https://a'", want: true},
		{name: "command substitution", command: "echo $(wget -qO- https://a)", want: true},
		{name: "safe command", command: "echo hello", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := detectHighRiskReason(tt.command)
			if got != tt.want {
				t.Fatalf("detectHighRiskReason(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}
