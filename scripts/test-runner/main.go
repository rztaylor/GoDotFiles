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
	TestRuns int
	Skipped  int
}

func main() {
	fmt.Println("Running tests...")

	expectedPackages, err := listPackages()
	if err != nil {
		fmt.Printf("Error listing packages: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "test", "-json", "-cover", "./...")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe: %v\n", err)
		os.Exit(1)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error creating stderr pipe: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting go test: %v\n", err)
		os.Exit(1)
	}

	results := make(map[string]*PackageResult)
	for _, pkg := range expectedPackages {
		results[pkg] = &PackageResult{Name: pkg, Status: "PENDING"}
	}

	startTime := time.Now()
	totalTests := 0
	consume := func(r *bufio.Scanner, echo bool) error {
		// Increase buffer for large JSON lines emitted by go test.
		buf := make([]byte, 0, 64*1024)
		r.Buffer(buf, 1024*1024)
		for r.Scan() {
			text := r.Text()
			if !strings.HasPrefix(text, "{") {
				if pkg := parseNoTestPackage(text); pkg != "" {
					res := ensureResult(results, pkg)
					res.Status = "NO_TEST"
				}
				if echo {
					fmt.Println(text)
				}
				continue
			}

			var event TestEvent
			if err := json.Unmarshal([]byte(text), &event); err != nil {
				if echo {
					fmt.Println(text)
				}
				continue
			}
			if event.Package == "" {
				continue
			}
			pkgRes := ensureResult(results, event.Package)

			if event.Action == "output" {
				if pkg := parseNoTestPackage(event.Output); pkg != "" {
					res := ensureResult(results, pkg)
					res.Status = "NO_TEST"
				}
				if cov := parseCoverage(event.Output); cov != "" {
					pkgRes.Coverage = cov
				}
			}

			if event.Test != "" {
				switch event.Action {
				case "run":
					totalTests++
					pkgRes.TestRuns++
				case "skip":
					pkgRes.Skipped++
				case "fail":
					pkgRes.Failures = append(pkgRes.Failures, fmt.Sprintf("FAIL: %s (%2fs)", event.Test, event.Elapsed))
				}
				continue
			}

			switch event.Action {
			case "pass":
				if pkgRes.Status != "NO_TEST" {
					pkgRes.Status = "PASS"
				}
				pkgRes.Elapsed = event.Elapsed
			case "fail":
				pkgRes.Status = "FAIL"
				pkgRes.Elapsed = event.Elapsed
			case "skip":
				if pkgRes.Status == "PENDING" {
					pkgRes.Status = "SKIP"
				}
			}
		}
		return r.Err()
	}

	if err := consume(bufio.NewScanner(stdout), false); err != nil {
		fmt.Printf("Error reading go test stdout: %v\n", err)
		os.Exit(1)
	}
	if err := consume(bufio.NewScanner(stderr), true); err != nil {
		fmt.Printf("Error reading go test stderr: %v\n", err)
		os.Exit(1)
	}

	waitErr := cmd.Wait()
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
	reported := 0
	pending := 0
	for _, pkgName := range packages {
		res := results[pkgName]
		switch res.Status {
		case "FAIL":
			hasFailures = true
			reported++
			fmt.Printf("❌ %-40s  %6.2fs  coverage: %s\n", res.Name, res.Elapsed, res.Coverage)
			for _, fail := range res.Failures {
				fmt.Printf("    %s\n", fail)
			}
		case "PASS":
			reported++
			fmt.Printf("✅ %-40s  %6.2fs  coverage: %s\n", res.Name, res.Elapsed, res.Coverage)
		case "NO_TEST":
			reported++
			fmt.Printf("➖ %-40s  no test files\n", res.Name)
		case "SKIP":
			reported++
			fmt.Printf("⏭️  %-40s  skipped\n", res.Name)
		default:
			pending++
		}
	}

	if waitErr != nil {
		hasFailures = true
		fmt.Printf("\n❌ go test process failed: %v\n", waitErr)
	}

	if hasFailures {
		fmt.Printf("\n❌ Tests Failed (%d tests across %d/%d packages, %s)\n", totalTests, reported, len(packages), totalDuration.Round(time.Millisecond))
		os.Exit(1)
	} else {
		fmt.Printf("\n✅ All Tests Passed (%d tests across %d/%d packages, %s)\n", totalTests, reported, len(packages), totalDuration.Round(time.Millisecond))
		if pending > 0 {
			fmt.Printf("⚠️  %d package(s) had no final status event; check go test output above.\n", pending)
		}
		os.Exit(0)
	}
}

func ensureResult(results map[string]*PackageResult, pkg string) *PackageResult {
	if _, exists := results[pkg]; !exists {
		results[pkg] = &PackageResult{Name: pkg, Status: "PENDING"}
	}
	return results[pkg]
}

func parseNoTestPackage(line string) string {
	if !strings.Contains(line, "[no test files]") {
		return ""
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return ""
	}
	for _, f := range fields {
		if strings.Contains(f, "/") && !strings.HasPrefix(f, "[") {
			return f
		}
	}
	return ""
}

func parseCoverage(output string) string {
	if !strings.Contains(output, "coverage:") {
		return ""
	}
	parts := strings.Fields(output)
	for i, p := range parts {
		if p == "coverage:" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func listPackages() ([]string, error) {
	out, err := exec.Command("go", "list", "./...").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	pkgs := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		pkgs = append(pkgs, strings.TrimSpace(line))
	}
	sort.Strings(pkgs)
	return pkgs, nil
}
