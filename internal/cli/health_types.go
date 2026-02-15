package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

type healthSeverity string

const (
	healthSeverityInfo    healthSeverity = "info"
	healthSeverityWarning healthSeverity = "warning"
	healthSeverityError   healthSeverity = "error"
)

type healthFinding struct {
	Code     string         `json:"code"`
	Severity healthSeverity `json:"severity"`
	Title    string         `json:"title"`
	Detail   string         `json:"detail,omitempty"`
	Path     string         `json:"path,omitempty"`
	Hint     string         `json:"hint,omitempty"`
}

type healthReport struct {
	Command  string          `json:"command"`
	OK       bool            `json:"ok"`
	Errors   int             `json:"errors"`
	Warnings int             `json:"warnings"`
	Info     int             `json:"info"`
	Findings []healthFinding `json:"findings"`
}

func (r *healthReport) add(f healthFinding) {
	r.Findings = append(r.Findings, f)
	switch f.Severity {
	case healthSeverityError:
		r.Errors++
	case healthSeverityWarning:
		r.Warnings++
	default:
		r.Info++
	}
	r.OK = r.Errors == 0
}

func (r *healthReport) sort() {
	sort.SliceStable(r.Findings, func(i, j int) bool {
		if r.Findings[i].Severity != r.Findings[j].Severity {
			return severityRank(r.Findings[i].Severity) < severityRank(r.Findings[j].Severity)
		}
		if r.Findings[i].Code != r.Findings[j].Code {
			return r.Findings[i].Code < r.Findings[j].Code
		}
		return r.Findings[i].Path < r.Findings[j].Path
	})
}

func severityRank(s healthSeverity) int {
	switch s {
	case healthSeverityError:
		return 0
	case healthSeverityWarning:
		return 1
	default:
		return 2
	}
}

func writeHealthJSON(w io.Writer, report *healthReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func writeHealthText(w io.Writer, report *healthReport) {
	if len(report.Findings) == 0 {
		fmt.Fprintln(w, "No issues found.")
		return
	}

	for _, f := range report.Findings {
		fmt.Fprintf(w, "[%s] %s: %s\n", f.Severity, f.Code, f.Title)
		if f.Path != "" {
			fmt.Fprintf(w, "  path: %s\n", f.Path)
		}
		if f.Detail != "" {
			fmt.Fprintf(w, "  detail: %s\n", f.Detail)
		}
		if f.Hint != "" {
			fmt.Fprintf(w, "  hint: %s\n", f.Hint)
		}
	}
	fmt.Fprintf(w, "\nSummary: %d error(s), %d warning(s), %d info\n", report.Errors, report.Warnings, report.Info)
}
