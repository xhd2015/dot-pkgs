package scan_repo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RepoIndex is the durable per-universe repo list (schema v1).
// On disk: <cacheRoot>/<universe>/repos.json
type RepoIndex struct {
	Version   int              `json:"version"`
	Universe  string           `json:"universe"` // "home" | "root"
	Base      string           `json:"base"`
	UpdatedAt string           `json:"updated_at"`
	Repos     []RepoIndexEntry `json:"repos"`
}

// RepoIndexEntry is one discovered checkout row in the universe index.
type RepoIndexEntry struct {
	Path     string `json:"path"`
	RepoType string `json:"repo_type"`
	GitDir   string `json:"git_dir"`
	Depth    int    `json:"depth"`
	SeenAt   string `json:"seen_at"`
}

// RepoIndexPath returns <cacheRoot>/<universe>/repos.json.
func RepoIndexPath(cacheRoot, universe string) string {
	return filepath.Join(cacheRoot, universe, "repos.json")
}

// SaveRepoIndex persists index under <cacheRoot>/<index.Universe>/repos.json
// with an atomic temp+rename write. Parent directories are created as needed.
func SaveRepoIndex(cacheRoot string, index RepoIndex) error {
	if index.Universe == "" {
		return fmt.Errorf("SaveRepoIndex: empty universe")
	}
	path := RepoIndexPath(cacheRoot, index.Universe)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(index)
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

// LoadRepoIndex reads <cacheRoot>/<universe>/repos.json.
// Missing file returns (zero, false, nil).
func LoadRepoIndex(cacheRoot, universe string) (RepoIndex, bool, error) {
	path := RepoIndexPath(cacheRoot, universe)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return RepoIndex{}, false, nil
		}
		return RepoIndex{}, false, err
	}
	var index RepoIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return RepoIndex{}, false, err
	}
	return index, true, nil
}

// ApplyLiveness returns a copy of index keeping only entries whose
// path/.git still exists (directory or gitfile). Envelope fields are preserved.
func ApplyLiveness(index RepoIndex) RepoIndex {
	out := index
	if index.Repos == nil {
		out.Repos = nil
		return out
	}
	kept := make([]RepoIndexEntry, 0, len(index.Repos))
	for _, e := range index.Repos {
		// Live if path/.git exists as a directory or gitfile.
		if _, err := os.Stat(filepath.Join(e.Path, ".git")); err == nil {
			kept = append(kept, e)
		}
	}
	out.Repos = kept
	return out
}
