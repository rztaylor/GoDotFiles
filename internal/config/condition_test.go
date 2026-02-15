package config

import (
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/platform"
)

func TestEvaluateCondition(t *testing.T) {
	p := &platform.Platform{
		OS:       "linux",
		Distro:   "ubuntu",
		Hostname: "work-laptop",
		Arch:     "amd64",
	}

	tests := []struct {
		name      string
		condition string
		want      bool
		wantErr   bool
	}{
		{
			name:      "simple equality",
			condition: "os == linux",
			want:      true,
		},
		{
			name:      "simple regex",
			condition: "hostname =~ '^work-.*'",
			want:      true,
		},
		{
			name:      "and condition",
			condition: "os == linux AND arch == amd64",
			want:      true,
		},
		{
			name:      "or condition",
			condition: "os == macos OR os == linux",
			want:      true,
		},
		{
			name:      "and takes precedence over or",
			condition: "os == macos OR os == linux AND arch == arm64",
			want:      false,
		},
		{
			name:      "parentheses change precedence",
			condition: "(os == macos OR os == linux) AND arch == amd64",
			want:      true,
		},
		{
			name:      "lowercase operators supported",
			condition: "os == linux and arch == amd64 or distro == fedora",
			want:      true,
		},
		{
			name:      "not macos via inequality",
			condition: "os != macos",
			want:      true,
		},
		{
			name:      "invalid expression",
			condition: "os ==",
			wantErr:   true,
		},
		{
			name:      "unknown field",
			condition: "kernel == linux",
			wantErr:   true,
		},
		{
			name:      "unbalanced parentheses",
			condition: "(os == linux OR arch == amd64",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateCondition(tt.condition, p)
			if tt.wantErr {
				if err == nil {
					t.Fatal("EvaluateCondition() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("EvaluateCondition() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}
