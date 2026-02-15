package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SnapshotCandidate represents a historical snapshot option for a target path.
type SnapshotCandidate struct {
	Target         string
	SnapshotPath   string
	SnapshotKind   string
	SnapshotMode   string
	LinkTarget     string
	CapturedAt     time.Time
	OperationAt    time.Time
	OperationLog   string
	OperationIndex int
}

// RollbackResult summarizes a rollback run.
type RollbackResult struct {
	Restored int
	Removed  int
	Failed   []string
}

// LatestOperationLog returns the newest operation log path and parsed operations.
func LatestOperationLog(gdfDir string) (string, []Operation, error) {
	logs, err := ListOperationLogs(gdfDir)
	if err != nil {
		return "", nil, err
	}
	if len(logs) == 0 {
		return "", nil, nil
	}
	latest := logs[len(logs)-1]
	ops, err := LoadOperationLog(latest)
	if err != nil {
		return "", nil, err
	}
	return latest, ops, nil
}

// ListOperationLogs returns operation log files in ascending order.
func ListOperationLogs(gdfDir string) ([]string, error) {
	logDir := filepath.Join(gdfDir, ".operations")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading operation logs: %w", err)
	}
	paths := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		paths = append(paths, filepath.Join(logDir, e.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}

// LoadOperationLog parses a JSON operation log.
func LoadOperationLog(path string) ([]Operation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading operation log %s: %w", path, err)
	}
	var ops []Operation
	if err := json.Unmarshal(data, &ops); err != nil {
		return nil, fmt.Errorf("parsing operation log %s: %w", path, err)
	}
	return ops, nil
}

// FindSnapshotCandidates finds all logged snapshots for a target across history.
func FindSnapshotCandidates(gdfDir, target string) ([]SnapshotCandidate, error) {
	logs, err := ListOperationLogs(gdfDir)
	if err != nil {
		return nil, err
	}

	var out []SnapshotCandidate
	for _, logPath := range logs {
		ops, err := LoadOperationLog(logPath)
		if err != nil {
			continue
		}
		for i, op := range ops {
			if op.Type != "link" || op.Target != target || op.Details == nil {
				continue
			}
			snapshotPath := op.Details["snapshot_path"]
			if snapshotPath == "" {
				continue
			}
			capturedAt := op.Timestamp
			if ts := op.Details["snapshot_captured_at"]; ts != "" {
				if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
					capturedAt = parsed
				}
			}
			out = append(out, SnapshotCandidate{
				Target:         target,
				SnapshotPath:   snapshotPath,
				SnapshotKind:   op.Details["snapshot_kind"],
				SnapshotMode:   op.Details["snapshot_mode"],
				LinkTarget:     op.Details["snapshot_link_target"],
				CapturedAt:     capturedAt,
				OperationAt:    op.Timestamp,
				OperationLog:   logPath,
				OperationIndex: i,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CapturedAt.After(out[j].CapturedAt)
	})
	return out, nil
}

// RollbackOperations reverts supported operations in reverse order.
func RollbackOperations(gdfDir string, ops []Operation, selector func(target string, candidates []SnapshotCandidate) (*SnapshotCandidate, error)) RollbackResult {
	result := RollbackResult{Failed: make([]string, 0)}
	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i]
		switch op.Type {
		case "link":
			if err := rollbackLink(gdfDir, op, selector); err != nil {
				result.Failed = append(result.Failed, fmt.Sprintf("%s: %v", op.Target, err))
				continue
			}
			if op.Details != nil && op.Details["snapshot_path"] != "" {
				result.Restored++
			} else {
				result.Removed++
			}
		}
	}
	return result
}

func rollbackLink(gdfDir string, op Operation, selector func(target string, candidates []SnapshotCandidate) (*SnapshotCandidate, error)) error {
	if op.Details == nil {
		return removeSymlinkIfManaged(op.Target, "")
	}

	candidate := &SnapshotCandidate{
		Target:       op.Target,
		SnapshotPath: op.Details["snapshot_path"],
		SnapshotKind: op.Details["snapshot_kind"],
		SnapshotMode: op.Details["snapshot_mode"],
		LinkTarget:   op.Details["snapshot_link_target"],
		CapturedAt:   op.Timestamp,
	}
	if ts := op.Details["snapshot_captured_at"]; ts != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			candidate.CapturedAt = parsed
		}
	}

	if candidate.SnapshotPath == "" {
		return removeSymlinkIfManaged(op.Target, op.Details["source_abs"])
	}

	if selector != nil {
		candidates, err := FindSnapshotCandidates(gdfDir, op.Target)
		if err == nil && len(candidates) > 1 {
			selected, err := selector(op.Target, candidates)
			if err != nil {
				return err
			}
			if selected != nil {
				candidate = selected
			}
		}
	}

	return restoreSnapshot(op.Target, candidate)
}

func removeSymlinkIfManaged(target, expectedDest string) error {
	info, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}
	if expectedDest != "" {
		dest, err := os.Readlink(target)
		if err != nil {
			return err
		}
		absDest := dest
		if !filepath.IsAbs(dest) {
			absDest = filepath.Clean(filepath.Join(filepath.Dir(target), dest))
		}
		if absDest != expectedDest {
			return nil
		}
	}
	return os.Remove(target)
}

func restoreSnapshot(target string, candidate *SnapshotCandidate) error {
	if candidate == nil || candidate.SnapshotPath == "" {
		return fmt.Errorf("missing snapshot candidate")
	}

	if _, err := os.Stat(candidate.SnapshotPath); err != nil {
		return fmt.Errorf("snapshot not found: %s", candidate.SnapshotPath)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("creating target dir: %w", err)
	}
	if err := os.RemoveAll(target); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing current target: %w", err)
	}

	switch candidate.SnapshotKind {
	case "symlink":
		linkTarget := candidate.LinkTarget
		if linkTarget == "" {
			data, err := os.ReadFile(candidate.SnapshotPath)
			if err != nil {
				return err
			}
			linkTarget = string(data)
		}
		return os.Symlink(linkTarget, target)
	case "file", "":
		in, err := os.Open(candidate.SnapshotPath)
		if err != nil {
			return err
		}
		defer in.Close()

		mode := os.FileMode(0644)
		if candidate.SnapshotMode != "" {
			trimmed := strings.TrimPrefix(candidate.SnapshotMode, "0")
			if trimmed == "" {
				trimmed = candidate.SnapshotMode
			}
			if parsed, err := strconv.ParseUint(trimmed, 8, 32); err == nil {
				mode = os.FileMode(parsed)
			}
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	default:
		return fmt.Errorf("unsupported snapshot kind: %s", candidate.SnapshotKind)
	}
}
