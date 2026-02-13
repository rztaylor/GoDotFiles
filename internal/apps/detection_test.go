package apps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectAppFromCommandIfExists(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		existing []string // app YAML files to create
		want     string
	}{
		{
			name:     "matches existing app",
			cmd:      "kubectl get pods",
			existing: []string{"kubectl"},
			want:     "kubectl",
		},
		{
			name:     "matches existing git app",
			cmd:      "git status",
			existing: []string{"git", "zsh"},
			want:     "git",
		},
		{
			name:     "no match when app does not exist",
			cmd:      "ls -la",
			existing: []string{"git", "kubectl"},
			want:     "",
		},
		{
			name:     "no match for shell builtin",
			cmd:      "cd ..",
			existing: []string{"git"},
			want:     "",
		},
		{
			name:     "no match for pipeline starting with cat",
			cmd:      "cat /tmp/file | jq .",
			existing: []string{"jq"},
			want:     "",
		},
		{
			name:     "matches absolute path command",
			cmd:      "/usr/bin/git log",
			existing: []string{"git"},
			want:     "git",
		},
		{
			name:     "empty command",
			cmd:      "",
			existing: []string{"git"},
			want:     "",
		},
		{
			name:     "no existing apps at all",
			cmd:      "kubectl get pods",
			existing: nil,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appsDir := t.TempDir()

			// Create dummy app YAML files
			for _, app := range tt.existing {
				path := filepath.Join(appsDir, app+".yaml")
				if err := os.WriteFile(path, []byte("name: "+app+"\n"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			got := DetectAppFromCommandIfExists(tt.cmd, appsDir)
			if got != tt.want {
				t.Errorf("DetectAppFromCommandIfExists(%q) = %q, want %q", tt.cmd, got, tt.want)
			}
		})
	}
}
