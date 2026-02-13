package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// TestEvent represents a line of JSON output from go test -json
type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test,omitempty"`
	Output  string    `json:"Output,omitempty"`
	Elapsed float64   `json:"Elapsed,omitempty"`
}

type PackageResult struct {
	Name     string
	Status   string
	Coverage string
	Elapsed  float64
	Failures []string
}

func main() {
	fmt.Println("Running tests...")

	cmd := exec.Command("go", "test", "-json", "-cover", "./...")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting go test: %v\n", err)
		os.Exit(1)
	}

	results := make(map[string]*PackageResult)
	scanner := bufio.NewScanner(stdout)

	// We need a large buffer for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	startTime := time.Now()
	totalTests := 0

	for scanner.Scan() {
		var event TestEvent
		text := scanner.Text()

		// Skip non-JSON lines (shouldn't happen with -json but good safety)
		if !strings.HasPrefix(text, "{") {
			fmt.Print(text) // Pass through raw output just in case
			continue
		}

		if err := json.Unmarshal([]byte(text), &event); err != nil {
			continue
		}

		if _, exists := results[event.Package]; !exists {
			results[event.Package] = &PackageResult{Name: event.Package, Status: "PENDING"}
		}
		pkgRes := results[event.Package]

		// Capture coverage output
		if event.Action == "output" {
			if strings.Contains(event.Output, "coverage:") {
				parts := strings.Fields(event.Output)
				for i, p := range parts {
					if p == "coverage:" && i+1 < len(parts) {
						pkgRes.Coverage = parts[i+1]
					}
				}
			}
		}

		if event.Test != "" {
			// Test level event
			// We count "run" events to get total tests started.
			// Alternatively count PASS/FAIL but skipping parallel subtests might be tricky.
			// "run" is reliable for "how many tests were attempted".
			if event.Action == "run" {
				totalTests++
			} else if event.Action == "fail" {
				pkgRes.Failures = append(pkgRes.Failures, fmt.Sprintf("FAIL: %s (%2fs)", event.Test, event.Elapsed))
			}
		} else {
			// Package level event
			if event.Action == "pass" {
				pkgRes.Status = "PASS"
				pkgRes.Elapsed = event.Elapsed
			} else if event.Action == "fail" {
				pkgRes.Status = "FAIL"
				pkgRes.Elapsed = event.Elapsed
			}
		}
	}

	cmd.Wait() // ignore error here, we track status via events
	totalDuration := time.Since(startTime)

	// Print Summary
	fmt.Println("\n=== Test Summary ===")

	// Sort packages
	packages := make([]string, 0, len(results))
	for p := range results {
		packages = append(packages, p)
	}
	sort.Strings(packages)

	hasFailures := false
	for _, pkgName := range packages {
		res := results[pkgName]
		if res.Status == "FAIL" {
			hasFailures = true
			fmt.Printf("❌ %-40s  %6.2fs  coverage: %s\n", res.Name, res.Elapsed, res.Coverage)
			for _, fail := range res.Failures {
				fmt.Printf("    %s\n", fail)
			}
		} else if res.Status == "PASS" {
			fmt.Printf("✅ %-40s  %6.2fs  coverage: %s\n", res.Name, res.Elapsed, res.Coverage)
		}
	}

	if hasFailures {
		fmt.Printf("\n❌ Tests Failed (%d tests, %s)\n", totalTests, totalDuration.Round(time.Millisecond))
		os.Exit(1)
	} else {
		fmt.Printf("\n✅ All Tests Passed (%d tests, %s)\n", totalTests, totalDuration.Round(time.Millisecond))
		os.Exit(0)
	}
}
