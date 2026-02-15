package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// Snapshot captures a historical copy of a target before mutation.
type Snapshot struct {
	ID           string
	OriginalPath string
	Path         string
	Kind         string // "file" or "symlink"
	LinkTarget   string
	Mode         os.FileMode
	SizeBytes    int64
	Checksum     string
	CapturedAt   time.Time
}

// HistoryManager stores and evicts file snapshots.
type HistoryManager struct {
	Dir      string
	MaxBytes int64
}

// NewHistoryManager creates a manager rooted at ~/.gdf/.history.
func NewHistoryManager(gdfDir string, maxSizeMB int) *HistoryManager {
	if maxSizeMB <= 0 {
		maxSizeMB = 512
	}
	return &HistoryManager{
		Dir:      filepath.Join(gdfDir, ".history"),
		MaxBytes: int64(maxSizeMB) * 1024 * 1024,
	}
}

// Capture snapshots the current contents of path. Missing paths return nil, nil.
func (h *HistoryManager) Capture(path string) (*Snapshot, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat target for snapshot: %w", err)
	}

	if info.Mode()&os.ModeSymlink == 0 && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("unsupported snapshot target mode for %s", path)
	}

	if err := os.MkdirAll(h.Dir, 0755); err != nil {
		return nil, fmt.Errorf("create history directory: %w", err)
	}

	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	snapshotPath := filepath.Join(h.Dir, id+".snap")
	capturedAt := time.Now().UTC()

	s := &Snapshot{
		ID:           id,
		OriginalPath: path,
		Path:         snapshotPath,
		CapturedAt:   capturedAt,
	}

	if info.Mode()&os.ModeSymlink != 0 {
		dest, err := os.Readlink(path)
		if err != nil {
			return nil, fmt.Errorf("reading symlink target: %w", err)
		}
		if err := os.WriteFile(snapshotPath, []byte(dest), 0644); err != nil {
			return nil, fmt.Errorf("writing symlink snapshot: %w", err)
		}
		s.Kind = "symlink"
		s.LinkTarget = dest
		s.SizeBytes = int64(len(dest))
		s.Mode = 0777
		hash := sha256.Sum256([]byte(dest))
		s.Checksum = hex.EncodeToString(hash[:])
	} else {
		sum, size, mode, err := copyFileWithChecksum(path, snapshotPath)
		if err != nil {
			return nil, fmt.Errorf("copying file snapshot: %w", err)
		}
		s.Kind = "file"
		s.SizeBytes = size
		s.Mode = mode
		s.Checksum = sum
	}

	if err := h.enforceQuota(snapshotPath); err != nil {
		return nil, err
	}

	return s, nil
}

func copyFileWithChecksum(src, dst string) (string, int64, os.FileMode, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", 0, 0, err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return "", 0, 0, err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return "", 0, 0, err
	}
	defer out.Close()

	h := sha256.New()
	w := io.MultiWriter(out, h)
	n, err := io.Copy(w, in)
	if err != nil {
		return "", 0, 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, info.Mode(), nil
}

func (h *HistoryManager) enforceQuota(protectedPath string) error {
	entries, err := os.ReadDir(h.Dir)
	if err != nil {
		return fmt.Errorf("read history directory: %w", err)
	}

	type item struct {
		path    string
		modTime time.Time
		size    int64
	}
	items := make([]item, 0, len(entries))
	var total int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return fmt.Errorf("stat history file: %w", err)
		}
		p := filepath.Join(h.Dir, e.Name())
		items = append(items, item{path: p, modTime: info.ModTime(), size: info.Size()})
		total += info.Size()
	}

	if total <= h.MaxBytes {
		return nil
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].modTime.Before(items[j].modTime)
	})

	for _, it := range items {
		if total <= h.MaxBytes {
			break
		}
		if it.path == protectedPath {
			continue
		}
		if err := os.Remove(it.path); err != nil {
			return fmt.Errorf("evicting old snapshot %s: %w", it.path, err)
		}
		total -= it.size
	}

	return nil
}
