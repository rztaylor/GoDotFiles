package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		bundles []*apps.Bundle
		shell   ShellType
		want    []string // Strings that should appear in output
		wantErr bool
	}{
		{
			name: "single bundle with aliases",
			bundles: []*apps.Bundle{
				{
					Name: "kubectl",
					Shell: &apps.Shell{
						Aliases: map[string]string{
							"k":   "kubectl",
							"kgp": "kubectl get pods",
						},
					},
				},
			},
			shell: Bash,
			want: []string{
				"alias k='kubectl'",
				"alias kgp='kubectl get pods'",
			},
			wantErr: false,
		},
		{
			name: "multiple bundles with aliases",
			bundles: []*apps.Bundle{
				{
					Name: "kubectl",
					Shell: &apps.Shell{
						Aliases: map[string]string{
							"k": "kubectl",
						},
					},
				},
				{
					Name: "git",
					Shell: &apps.Shell{
						Aliases: map[string]string{
							"g":   "git",
							"gst": "git status",
						},
					},
				},
			},
			shell: Bash,
			want: []string{
				"alias k='kubectl'",
				"alias g='git'",
				"alias gst='git status'",
			},
			wantErr: false,
		},
		{
			name: "environment variables",
			bundles: []*apps.Bundle{
				{
					Name: "kubectl",
					Shell: &apps.Shell{
						Env: map[string]string{
							"KUBECONFIG": "$HOME/.kube/config",
							"EDITOR":     "vim",
						},
					},
				},
			},
			shell: Bash,
			want: []string{
				"export KUBECONFIG=",
				"export EDITOR=",
			},
			wantErr: false,
		},
		{
			name: "shell functions",
			bundles: []*apps.Bundle{
				{
					Name: "kubectl",
					Shell: &apps.Shell{
						Functions: map[string]string{
							"kns": "kubectl config set-context --current --namespace=\"$1\"",
						},
					},
				},
			},
			shell: Bash,
			want: []string{
				"kns()",
				"kubectl config set-context",
			},
			wantErr: false,
		},
		{
			name: "completions bash",
			bundles: []*apps.Bundle{
				{
					Name: "kubectl",
					Shell: &apps.Shell{
						Completions: &apps.Completions{
							Bash: "kubectl completion bash",
						},
					},
				},
			},
			shell: Bash,
			want: []string{
				"kubectl completion bash",
			},
			wantErr: false,
		},
		{
			name: "completions zsh",
			bundles: []*apps.Bundle{
				{
					Name: "kubectl",
					Shell: &apps.Shell{
						Completions: &apps.Completions{
							Zsh: "kubectl completion zsh",
						},
					},
				},
			},
			shell: Zsh,
			want: []string{
				"kubectl completion zsh",
			},
			wantErr: false,
		},
		{
			name: "combined - aliases, env, functions, completions",
			bundles: []*apps.Bundle{
				{
					Name: "kubectl",
					Shell: &apps.Shell{
						Aliases: map[string]string{
							"k": "kubectl",
						},
						Env: map[string]string{
							"KUBECONFIG": "$HOME/.kube/config",
						},
						Functions: map[string]string{
							"kns": "kubectl config set-context --current --namespace=\"$1\"",
						},
						Completions: &apps.Completions{
							Bash: "kubectl completion bash",
						},
					},
				},
			},
			shell: Bash,
			want: []string{
				"alias k='kubectl'",
				"export KUBECONFIG=",
				"kns()",
				"kubectl completion bash",
			},
			wantErr: false,
		},
		{
			name:    "empty bundles",
			bundles: []*apps.Bundle{},
			shell:   Bash,
			want:    []string{"#!/bin/bash"}, // At least header
			wantErr: false,
		},
		{
			name: "bundles without shell config",
			bundles: []*apps.Bundle{
				{
					Name:  "no-shell",
					Shell: nil,
				},
			},
			shell:   Bash,
			want:    []string{"#!/bin/bash"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "init.sh")

			g := NewGenerator()
			err := g.Generate(tt.bundles, tt.shell, outputPath, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Read generated file
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			contentStr := string(content)

			// Check for expected strings
			for _, want := range tt.want {
				if !strings.Contains(contentStr, want) {
					t.Errorf("Generated content missing %q.\nGot:\n%s", want, contentStr)
				}
			}
		})
	}
}

func TestGenerator_DuplicateAliases(t *testing.T) {
	// Test that last bundle wins for duplicate aliases
	bundles := []*apps.Bundle{
		{
			Name: "bundle1",
			Shell: &apps.Shell{
				Aliases: map[string]string{
					"k": "kubectl",
				},
			},
		},
		{
			Name: "bundle2",
			Shell: &apps.Shell{
				Aliases: map[string]string{
					"k": "k9s", // Override
				},
			},
		},
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "init.sh")

	g := NewGenerator()
	err := g.Generate(bundles, Bash, outputPath, nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Should contain the LAST definition
	if !strings.Contains(contentStr, "alias k='k9s'") {
		t.Errorf("Expected last alias definition to win. Got:\n%s", contentStr)
	}
}

func TestGenerator_ShellSpecific(t *testing.T) {
	bundles := []*apps.Bundle{
		{
			Name: "kubectl",
			Shell: &apps.Shell{
				Completions: &apps.Completions{
					Bash: "kubectl completion bash",
					Zsh:  "kubectl completion zsh",
				},
			},
		},
	}

	tests := []struct {
		name      string
		shellType ShellType
		wantShell string
		wantComp  string
	}{
		{
			name:      "bash",
			shellType: Bash,
			wantShell: "#!/bin/bash",
			wantComp:  "kubectl completion bash",
		},
		{
			name:      "zsh",
			shellType: Zsh,
			wantShell: "#!/bin/zsh",
			wantComp:  "kubectl completion zsh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "init.sh")

			g := NewGenerator()
			err := g.Generate(bundles, tt.shellType, outputPath, nil)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			contentStr := string(content)

			if !strings.Contains(contentStr, tt.wantShell) {
				t.Errorf("Expected shell header %q, got:\n%s", tt.wantShell, contentStr)
			}

			if !strings.Contains(contentStr, tt.wantComp) {
				t.Errorf("Expected completion %q, got:\n%s", tt.wantComp, contentStr)
			}
		})
	}
}

func TestGenerator_InitSnippets(t *testing.T) {
	bundles := []*apps.Bundle{
		{
			Name: "fnm",
			Shell: &apps.Shell{
				Init: []apps.InitSnippet{
					{
						Name:   "path",
						Common: `export PATH="$HOME/.local/share/fnm:$PATH"`,
					},
					{
						Name:  "env",
						Bash:  `eval "$(fnm env --shell bash)"`,
						Zsh:   `eval "$(fnm env --shell zsh)"`,
						Guard: "command -v fnm >/dev/null 2>&1",
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		shellType ShellType
		want      []string
		notWant   []string
	}{
		{
			name:      "bash picks bash-specific command",
			shellType: Bash,
			want: []string{
				"# Init",
				"# fnm:path",
				`export PATH="$HOME/.local/share/fnm:$PATH"`,
				"# fnm:env",
				"if command -v fnm >/dev/null 2>&1; then",
				`eval "$(fnm env --shell bash)"`,
			},
			notWant: []string{
				`eval "$(fnm env --shell zsh)"`,
			},
		},
		{
			name:      "zsh picks zsh-specific command",
			shellType: Zsh,
			want: []string{
				`eval "$(fnm env --shell zsh)"`,
			},
			notWant: []string{
				`eval "$(fnm env --shell bash)"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "init.sh")

			g := NewGenerator()
			err := g.Generate(bundles, tt.shellType, outputPath, nil)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			contentStr := string(content)
			for _, want := range tt.want {
				if !strings.Contains(contentStr, want) {
					t.Errorf("Generated content missing %q.\nGot:\n%s", want, contentStr)
				}
			}
			for _, notWant := range tt.notWant {
				if strings.Contains(contentStr, notWant) {
					t.Errorf("Generated content unexpectedly contains %q.\nGot:\n%s", notWant, contentStr)
				}
			}
		})
	}
}

func TestGenerator_InitOrdering(t *testing.T) {
	bundles := []*apps.Bundle{
		{
			Name: "base",
			Shell: &apps.Shell{
				Init: []apps.InitSnippet{
					{Name: "a", Common: "echo base-a"},
					{Name: "b", Common: "echo base-b"},
				},
			},
		},
		{
			Name: "work",
			Shell: &apps.Shell{
				Init: []apps.InitSnippet{
					{Name: "c", Common: "echo work-c"},
				},
			},
		},
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "init.sh")

	g := NewGenerator()
	err := g.Generate(bundles, Bash, outputPath, nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	first := strings.Index(contentStr, "echo base-a")
	second := strings.Index(contentStr, "echo base-b")
	third := strings.Index(contentStr, "echo work-c")

	if first == -1 || second == -1 || third == -1 {
		t.Fatalf("missing init snippets in generated output:\n%s", contentStr)
	}
	if !(first < second && second < third) {
		t.Errorf("init snippets are out of order:\n%s", contentStr)
	}
}
