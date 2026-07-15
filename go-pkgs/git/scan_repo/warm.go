package scan_repo

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	defaultWarmRefreshBudget = time.Second
	defaultYoungAge          = 60 * time.Second
)

// rootWarmEligible reports whether a scan root has a complete mirror entry
// that can be served without a full live WalkDir.
// Empty options_hash is treated as a match (P3); non-empty is also accepted
// until option-hash invalidation is implemented.
func rootWarmEligible(cacheRoot, absRoot string) bool {
	if cacheRoot == "" {
		return false
	}
	entry, ok, err := LoadCacheEntry(cacheRoot, absRoot)
	if err != nil || !ok {
		return false
	}
	return entry.ScanComplete
}

// warmServeRoot serves repos from mirror is_repo marks under absRoot, with
// liveness checks against the real filesystem. It does not WalkDir the live tree.
func warmServeRoot(ctx context.Context, absRoot, cacheRoot string, onRepo func(Repo) error) ([]Repo, error) {
	candidates, err := listCachedRepoPaths(cacheRoot, absRoot)
	if err != nil {
		return nil, err
	}

	var repos []Repo
	for _, path := range candidates {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		repo, live, err := liveRepoFromCache(cacheRoot, path)
		if err != nil {
			return nil, err
		}
		if !live {
			continue
		}
		if onRepo != nil {
			if err := onRepo(repo); err != nil {
				return nil, err
			}
		} else {
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

// listCachedRepoPaths walks the mirror tree under absRoot and returns real
// paths whose cache entries have is_repo=true.
func listCachedRepoPaths(cacheRoot, absRoot string) ([]string, error) {
	entryPath, err := MirrorEntryPath(cacheRoot, absRoot)
	if err != nil {
		return nil, err
	}
	mirrorDir := filepath.Dir(entryPath)
	mirrorRoot := filepath.Join(cacheRoot, "mirror")

	var paths []string
	err = filepath.WalkDir(mirrorDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() || d.Name() != "entry.json" {
			return nil
		}
		rel, relErr := filepath.Rel(mirrorRoot, filepath.Dir(p))
		if relErr != nil {
			return relErr
		}
		// Reconstruct absolute real path from mirror relative segments.
		realPath := filepath.Clean(string(filepath.Separator) + rel)

		entry, ok, loadErr := LoadCacheEntry(cacheRoot, realPath)
		if loadErr != nil {
			return loadErr
		}
		if ok && entry.IsRepo {
			paths = append(paths, realPath)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return paths, nil
}

// liveRepoFromCache verifies path/.git still exists. If gone, clears the
// is_repo mark and returns live=false. If present, builds a Repo from the live
// .git (dir or gitlink).
func liveRepoFromCache(cacheRoot, path string) (Repo, bool, error) {
	gitPath := filepath.Join(path, ".git")
	info, statErr := os.Stat(gitPath)
	if statErr != nil || !(info.IsDir() || info.Mode().IsRegular()) {
		if err := clearRepoMark(cacheRoot, path); err != nil {
			return Repo{}, false, err
		}
		return Repo{}, false, nil
	}

	gitDir, repoType, resolveErr := resolveGitDir(path, gitPath, info)
	if resolveErr != nil {
		// Still a checkout of sorts; surface like cold walk.
		return Repo{
			Path:  path,
			Name:  filepath.Base(path),
			Error: resolveErr.Error(),
		}, true, nil
	}
	// parseGitLink failure yields empty gitDir with nil error; treat as non-repo.
	if gitDir == "" && repoType == "" {
		if err := clearRepoMark(cacheRoot, path); err != nil {
			return Repo{}, false, err
		}
		return Repo{}, false, nil
	}

	return Repo{
		Path:     path,
		Name:     filepath.Base(path),
		GitDir:   gitDir,
		RepoType: repoType,
	}, true, nil
}

// clearRepoMark sets is_repo=false on the mirror entry (or is a no-op if missing).
func clearRepoMark(cacheRoot, path string) error {
	entry, ok, err := LoadCacheEntry(cacheRoot, path)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if !entry.IsRepo {
		return nil
	}
	entry.IsRepo = false
	entry.RepoType = ""
	entry.GitDir = ""
	entry.RefreshedAt = time.Now().UTC().Format(time.RFC3339)
	return SaveCacheEntry(cacheRoot, path, entry)
}

// warmBudgetRefresh rewalks oldest eligible direct-child units under absRoot
// until WarmRefreshBudget wall time is exhausted, merging newly found repos
// into existing. Negative budget means no refresh work. Zero budget uses the
// product default (1s). Cold paths never call this.
//
// Units without a parseable mirror refreshed_at are not eligible (so brand-new
// root-level dirs planted after cold seed stay soft-incomplete under default
// YoungAge, matching P3 warm serve-only behavior).
func warmBudgetRefresh(ctx context.Context, absRoot string, opts Options, cacheRoot string, ignore ignoreConfig, existing []Repo, onRepo func(Repo) error) ([]Repo, error) {
	budget, ok := resolveWarmRefreshBudget(opts.WarmRefreshBudget)
	if !ok {
		return existing, nil
	}
	youngAge := resolveYoungAge(opts.YoungAge)
	now := resolveNow(opts)

	units, err := listRefreshUnits(cacheRoot, absRoot)
	if err != nil {
		return existing, err
	}
	// No direct-child units: optionally rewalk root once (design fallback).
	if len(units) == 0 {
		units = []string{absRoot}
	}

	candidates := eligibleRefreshUnits(cacheRoot, units, now, youngAge)
	if len(candidates) == 0 {
		return existing, nil
	}

	seen := make(map[string]struct{}, len(existing))
	for _, r := range existing {
		seen[r.Path] = struct{}{}
	}

	start := time.Now()
	for _, u := range candidates {
		if time.Since(start) >= budget {
			break
		}
		select {
		case <-ctx.Done():
			return existing, ctx.Err()
		default:
		}

		found, walkErr := walkRoot(ctx, u.path, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr, nil, cacheRoot)
		if walkErr != nil {
			return existing, walkErr
		}
		for _, repo := range found {
			if _, dup := seen[repo.Path]; dup {
				continue
			}
			seen[repo.Path] = struct{}{}
			if onRepo != nil {
				if err := onRepo(repo); err != nil {
					return existing, err
				}
			}
			existing = append(existing, repo)
		}
	}
	return existing, nil
}

// resolveWarmRefreshBudget returns (budget, enabled). Negative → disabled.
// Zero → default 1s. Positive → that duration.
func resolveWarmRefreshBudget(d time.Duration) (time.Duration, bool) {
	if d < 0 {
		return 0, false
	}
	if d == 0 {
		return defaultWarmRefreshBudget, true
	}
	return d, true
}

func resolveYoungAge(d time.Duration) time.Duration {
	if d == 0 {
		return defaultYoungAge
	}
	return d
}

func resolveNow(opts Options) time.Time {
	if opts.Now != nil {
		return opts.Now()
	}
	return time.Now()
}

// listRefreshUnits returns direct child directories of absRoot from the live
// filesystem unioned with children recorded on the root mirror entry.
func listRefreshUnits(cacheRoot, absRoot string) ([]string, error) {
	seen := make(map[string]struct{})
	var units []string
	add := func(p string) {
		p = filepath.Clean(p)
		if p == "" || p == absRoot {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		info, err := os.Stat(p)
		if err != nil || !info.IsDir() {
			return
		}
		seen[p] = struct{}{}
		units = append(units, p)
	}

	entries, err := os.ReadDir(absRoot)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			add(filepath.Join(absRoot, e.Name()))
		}
	}

	if cacheRoot != "" {
		if entry, ok, loadErr := LoadCacheEntry(cacheRoot, absRoot); loadErr != nil {
			return nil, loadErr
		} else if ok {
			for _, name := range entry.Children {
				if name == "" || name == "." || name == ".." {
					continue
				}
				add(filepath.Join(absRoot, name))
			}
		}
	}
	return units, nil
}

type refreshUnit struct {
	path string
	at   time.Time
}

// eligibleRefreshUnits keeps units with parseable refreshed_at at least youngAge
// old, sorted oldest-first (then path for stability).
func eligibleRefreshUnits(cacheRoot string, units []string, now time.Time, youngAge time.Duration) []refreshUnit {
	var out []refreshUnit
	for _, u := range units {
		entry, ok, err := LoadCacheEntry(cacheRoot, u)
		if err != nil || !ok {
			continue
		}
		at, err := time.Parse(time.RFC3339, entry.RefreshedAt)
		if err != nil {
			continue
		}
		if now.Sub(at) < youngAge {
			continue
		}
		out = append(out, refreshUnit{path: u, at: at})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].at.Equal(out[j].at) {
			return out[i].path < out[j].path
		}
		return out[i].at.Before(out[j].at)
	})
	return out
}
