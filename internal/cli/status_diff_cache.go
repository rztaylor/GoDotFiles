package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v3"
)

type driftCache struct {
	Entries map[string]driftCacheEntry `yaml:"entries"`
}

type driftCacheEntry struct {
	SourceSize  int64  `yaml:"source_size"`
	SourceMtime int64  `yaml:"source_mtime"`
	TargetSize  int64  `yaml:"target_size"`
	TargetMtime int64  `yaml:"target_mtime"`
	Preview     string `yaml:"preview"`
	Patch       string `yaml:"patch,omitempty"`
	PatchSkip   string `yaml:"patch_skip,omitempty"`
}

func loadDriftCache(gdfDir string) *driftCache {
	path := driftCachePath(gdfDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return &driftCache{Entries: map[string]driftCacheEntry{}}
	}
	var cache driftCache
	if err := yaml.Unmarshal(data, &cache); err != nil || cache.Entries == nil {
		return &driftCache{Entries: map[string]driftCacheEntry{}}
	}
	return &cache
}

func saveDriftCache(gdfDir string, cache *driftCache) error {
	if cache == nil {
		return nil
	}
	data, err := yaml.Marshal(cache)
	if err != nil {
		return err
	}
	path := driftCachePath(gdfDir)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func driftCachePath(gdfDir string) string {
	return filepath.Join(gdfDir, ".cache", "status-diff.yaml")
}

func cachedDiffDetails(cache *driftCache, sourcePath, targetPath string, wantPatch bool, maxBytes int64) (preview, patch, patchSkip string, updated bool) {
	if cache == nil || cache.Entries == nil {
		cache = &driftCache{Entries: map[string]driftCacheEntry{}}
	}
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return "", "", "", false
	}
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		return "", "", "", false
	}

	key := sourcePath + "|" + targetPath
	entry, ok := cache.Entries[key]
	if ok &&
		entry.SourceSize == sourceInfo.Size() &&
		entry.SourceMtime == sourceInfo.ModTime().UnixNano() &&
		entry.TargetSize == targetInfo.Size() &&
		entry.TargetMtime == targetInfo.ModTime().UnixNano() &&
		(!wantPatch || entry.Patch != "" || entry.PatchSkip != "") {
		return entry.Preview, entry.Patch, entry.PatchSkip, false
	}

	preview = diffPreview(sourcePath, targetPath)
	if wantPatch {
		patch, patchSkip = unifiedPatch(sourcePath, targetPath, maxBytes)
	}

	cache.Entries[key] = driftCacheEntry{
		SourceSize:  sourceInfo.Size(),
		SourceMtime: sourceInfo.ModTime().UnixNano(),
		TargetSize:  targetInfo.Size(),
		TargetMtime: targetInfo.ModTime().UnixNano(),
		Preview:     preview,
		Patch:       patch,
		PatchSkip:   patchSkip,
	}
	return preview, patch, patchSkip, true
}

func unifiedPatch(sourcePath, targetPath string, maxBytes int64) (string, string) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return "", "source unavailable"
	}
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		return "", "target unavailable"
	}
	if sourceInfo.Size() > maxBytes || targetInfo.Size() > maxBytes {
		return "", fmt.Sprintf("file exceeds --max-bytes limit (%d)", maxBytes)
	}

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", "source read failed"
	}
	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return "", "target read failed"
	}
	if strings.ContainsRune(string(sourceData), '\x00') || strings.ContainsRune(string(targetData), '\x00') {
		return "", "binary content differs"
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(targetData)),
		B:        difflib.SplitLines(string(sourceData)),
		FromFile: targetPath,
		ToFile:   sourcePath,
		Context:  3,
	}
	out, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return "", "failed to build patch"
	}
	if strings.TrimSpace(out) == "" {
		return "", "content matches but target is not linked"
	}
	return out, ""
}
