package platform

import (
	"os"
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		want     string
	}{
		{
			name:     "bash",
			shellEnv: "/bin/bash",
			want:     "bash",
		},
		{
			name:     "zsh",
			shellEnv: "/usr/bin/zsh",
			want:     "zsh",
		},
		{
			name:     "fish",
			shellEnv: "/usr/local/bin/fish",
			want:     "fish",
		},
		{
			name:     "unknown shell",
			shellEnv: "/bin/ksh",
			want:     "unknown",
		},
		{
			name:     "empty SHELL",
			shellEnv: "",
			want:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original SHELL
			original := os.Getenv("SHELL")
			defer os.Setenv("SHELL", original)

			os.Setenv("SHELL", tt.shellEnv)

			got := DetectShell()
			if got != tt.want {
				t.Errorf("DetectShell() = %v, want %v", got, tt.want)
			}
		})
	}
}
