package scan_repo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CacheEntry is the v1 on-disk schema for a single directory under the mirror cache.
type CacheEntry struct {
	Version      int      `json:"version"`
	RefreshedAt  string   `json:"refreshed_at"`
	MtimeNs      int64    `json:"mtime_ns"`
	IsRepo       bool     `json:"is_repo"`
	RepoType     string   `json:"repo_type"`
	GitDir       string   `json:"git_dir"`
	Children     []string `json:"children"`
	ScanComplete bool     `json:"scan_complete"`
	OptionsHash  string   `json:"options_hash"`
}

// MirrorEntryPath maps a real directory path to its mirror entry.json path:
//
//	<cacheRoot>/mirror/<abs-without-leading-slash>/entry.json
func MirrorEntryPath(cacheRoot, realPath string) (string, error) {
	abs, err := filepath.Abs(realPath)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	rel := strings.TrimPrefix(abs, string(filepath.Separator))
	return filepath.Join(cacheRoot, "mirror", rel, "entry.json"), nil
}

// SaveCacheEntry writes entry as JSON under the mirror path for realPath.
// Intermediate directories are created as needed. The write is atomic
// (temp file + rename).
func SaveCacheEntry(cacheRoot, realPath string, entry CacheEntry) error {
	path, err := MirrorEntryPath(cacheRoot, realPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", path, os.Getpid())
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// LoadCacheEntry reads the mirror entry for realPath.
// Missing file returns (zero, false, nil). Corrupt JSON returns an error.
func LoadCacheEntry(cacheRoot, realPath string) (CacheEntry, bool, error) {
	path, err := MirrorEntryPath(cacheRoot, realPath)
	if err != nil {
		return CacheEntry{}, false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return CacheEntry{}, false, nil
		}
		return CacheEntry{}, false, err
	}
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return CacheEntry{}, false, err
	}
	return entry, true, nil
}
