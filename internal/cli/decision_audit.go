package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// decisionRecord captures an explicit conflict-resolution choice for auditing.
type decisionRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command"`
	Subject   string    `json:"subject"`
	Conflict  string    `json:"conflict"`
	Choice    string    `json:"choice"`
}

// decisionAudit accumulates explicit user decisions and persists them under .operations/.
type decisionAudit struct {
	command  string
	records  []decisionRecord
	disabled bool
}

func newDecisionAudit(command string, disabled bool) *decisionAudit {
	return &decisionAudit{command: command, disabled: disabled}
}

func (d *decisionAudit) Record(subject, conflict, choice string) {
	if d == nil || d.disabled {
		return
	}
	d.records = append(d.records, decisionRecord{
		Timestamp: time.Now(),
		Command:   d.command,
		Subject:   subject,
		Conflict:  conflict,
		Choice:    choice,
	})
}

func (d *decisionAudit) Save(gdfDir string) (string, error) {
	if d == nil || d.disabled || len(d.records) == 0 {
		return "", nil
	}

	logDir := filepath.Join(gdfDir, ".operations")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", fmt.Errorf("creating decision log directory: %w", err)
	}

	name := fmt.Sprintf("decisions-%s.json", time.Now().Format("20060102-150405"))
	path := filepath.Join(logDir, name)

	data, err := json.MarshalIndent(d.records, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling decision log: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("writing decision log: %w", err)
	}
	return path, nil
}
