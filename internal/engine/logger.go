package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Operation represents a single operation performed during apply.
type Operation struct {
	Type      string            `json:"type"`      // "link", "package_install", "hook_run", "shell_generate"
	Timestamp time.Time         `json:"timestamp"` // When the operation was performed
	Target    string            `json:"target"`    // What was affected (e.g., symlink path, package name)
	Details   map[string]string `json:"details"`   // Additional context
}

// Logger records operations for potential rollback.
type Logger struct {
	operations []Operation
	dryRun     bool
}

// NewLogger creates a new operation logger.
func NewLogger(dryRun bool) *Logger {
	return &Logger{
		operations: make([]Operation, 0),
		dryRun:     dryRun,
	}
}

// Log records an operation.
func (l *Logger) Log(opType, target string, details map[string]string) {
	op := Operation{
		Type:      opType,
		Timestamp: time.Now(),
		Target:    target,
		Details:   details,
	}
	l.operations = append(l.operations, op)
}

// Save writes the operation log to a file.
// Returns the path where the log was saved, or error.
func (l *Logger) Save(gdfDir string) (string, error) {
	if l.dryRun {
		return "", nil // Don't save logs for dry runs
	}

	if len(l.operations) == 0 {
		return "", nil // Nothing to save
	}

	// Create .operations directory
	logDir := filepath.Join(gdfDir, ".operations")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", fmt.Errorf("creating log directory: %w", err)
	}

	// Generate log filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	logPath := filepath.Join(logDir, fmt.Sprintf("%s.json", timestamp))

	// Marshal operations to JSON
	data, err := json.MarshalIndent(l.operations, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling operations: %w", err)
	}

	// Write to file
	if err := os.WriteFile(logPath, data, 0644); err != nil {
		return "", fmt.Errorf("writing log file: %w", err)
	}

	return logPath, nil
}

// Operations returns all recorded operations.
func (l *Logger) Operations() []Operation {
	return l.operations
}

// IsDryRun returns whether this is a dry-run logger.
func (l *Logger) IsDryRun() bool {
	return l.dryRun
}
