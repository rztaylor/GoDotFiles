package shell

import (
	"testing"
)

func TestShellType_String(t *testing.T) {
	tests := []struct {
		name      string
		shellType ShellType
		want      string
	}{
		{
			name:      "bash",
			shellType: Bash,
			want:      "bash",
		},
		{
			name:      "zsh",
			shellType: Zsh,
			want:      "zsh",
		},
		{
			name:      "unknown",
			shellType: Unknown,
			want:      "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.shellType.String(); got != tt.want {
				t.Errorf("ShellType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseShellType(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		want  ShellType
	}{
		{
			name:  "parse bash",
			shell: "bash",
			want:  Bash,
		},
		{
			name:  "parse zsh",
			shell: "zsh",
			want:  Zsh,
		},
		{
			name:  "parse unknown",
			shell: "fish",
			want:  Unknown,
		},
		{
			name:  "parse empty",
			shell: "",
			want:  Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseShellType(tt.shell); got != tt.want {
				t.Errorf("ParseShellType() = %v, want %v", got, tt.want)
			}
		})
	}
}
