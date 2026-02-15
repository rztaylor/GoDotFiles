package main

import "testing"

func TestParseNoTestPackage(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "standard no test line",
			line: "?   \tgithub.com/rztaylor/GoDotFiles/cmd/gdf\t[no test files]",
			want: "github.com/rztaylor/GoDotFiles/cmd/gdf",
		},
		{
			name: "not a no-test line",
			line: "ok   \tgithub.com/rztaylor/GoDotFiles/internal/cli\t0.123s",
			want: "",
		},
		{
			name: "malformed line",
			line: "[no test files]",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNoTestPackage(tt.line)
			if got != tt.want {
				t.Fatalf("parseNoTestPackage(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}

func TestParseCoverage(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "coverage in output",
			output: "coverage: 73.2% of statements\n",
			want:   "73.2%",
		},
		{
			name:   "no coverage",
			output: "PASS\n",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCoverage(tt.output)
			if got != tt.want {
				t.Fatalf("parseCoverage(%q) = %q, want %q", tt.output, got, tt.want)
			}
		})
	}
}
