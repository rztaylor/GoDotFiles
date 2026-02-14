package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	file, err := os.Open("CHANGELOG.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening CHANGELOG.md: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var releaseNotes []string
	inLatestRelease := false

	// Regex to match version headers like "## [0.7.0] - 2026-02-14"
	versionHeaderRegex := regexp.MustCompile(`^## \[\d+\.\d+\.\d+\] - \d{4}-\d{2}-\d{2}`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if we found the latest release header
		if versionHeaderRegex.MatchString(line) {
			if inLatestRelease {
				// We found the NEXT release header, so we stop
				break
			}
			inLatestRelease = true
			releaseNotes = append(releaseNotes, line) // Include the header
			continue
		}

		if inLatestRelease {
			// Stop if we hit the bottom link references like "[Unreleased]: ..."
			if strings.HasPrefix(line, "[") && strings.Contains(line, "]:") {
				break
			}
			releaseNotes = append(releaseNotes, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	if len(releaseNotes) == 0 {
		fmt.Fprintln(os.Stderr, "No release notes found.")
		os.Exit(1)
	}

	// Print release notes to stdout
	fmt.Println(strings.TrimSpace(strings.Join(releaseNotes, "\n")))
}
