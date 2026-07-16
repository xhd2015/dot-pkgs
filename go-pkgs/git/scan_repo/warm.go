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
	// minBudgetForMidUnitCap is the smallest remaining budget for which a unit
	// rewalk gets a child-context deadline. Below this, a unit may still start
	// when remaining > 0 (between-unit gate) but runs under the parent context
	// so a single small unit can finish under a tiny positive budget used by
	// tests (e.g. 1µs oldest-first). Larger budgets (e.g. 100ms) always cap
	// mid-walk via child deadline + soft partial merge.
	minBudgetForMidUnitCap = time.Millisecond
)

// rootWarmEligible reports whether a scan root has a complete mirror entry
// that can be served without a full live WalkDir.
// Empty options_hash is treated as a match (P3); non-empty is also accepted
// until option-hash invalidation is implemented.
func rootWarmEligible(cacheRoot, absRoot string) bool {
	ok, _ := rootCacheMode(cacheRoot, absRoot, Options{})
	return ok
}

// rootCacheMode decides warm vs cold for a scan root and returns a greppable
// reason token for Debug logs (missing_root_entry, scan_complete_false,
// no_cache, refresh, ok).
func rootCacheMode(cacheRoot, absRoot string, opts Options) (warm bool, reason string) {
	if opts.NoCache || cacheRoot == "" {
		return false, "no_cache"
	}
	if opts.Refresh {
		return false, "refresh"
	}
	entry, ok, err := LoadCacheEntry(cacheRoot, absRoot)
	if err != nil || !ok {
		return false, "missing_root_entry"
	}
	if !entry.ScanComplete {
		return false, "scan_complete_false"
	}
	return true, "ok"
}

// warmServeStats is phase-level serve timing for Debug logs.
type warmServeStats struct {
	candidates int
	live       int
	duration   time.Duration
}

// warmServeRoot serves repos from mirror is_repo marks under absRoot, with
// liveness checks against the real filesystem. It does not WalkDir the live tree.
func warmServeRoot(ctx context.Context, absRoot, cacheRoot string, onRepo func(Repo) error) ([]Repo, warmServeStats, error) {
	start := time.Now()
	var stats warmServeStats

	candidates, err := listCachedRepoPaths(cacheRoot, absRoot)
	if err != nil {
		stats.duration = time.Since(start)
		return nil, stats, err
	}
	stats.candidates = len(candidates)

	var repos []Repo
	for _, path := range candidates {
		select {
		case <-ctx.Done():
			stats.duration = time.Since(start)
			return nil, stats, ctx.Err()
		default:
		}

		repo, live, err := liveRepoFromCache(cacheRoot, path)
		if err != nil {
			stats.duration = time.Since(start)
			return nil, stats, err
		}
		if !live {
			continue
		}
		stats.live++
		if onRepo != nil {
			if err := onRepo(repo); err != nil {
				stats.duration = time.Since(start)
				return nil, stats, err
			}
		} else {
			repos = append(repos, repo)
		}
	}
	stats.duration = time.Since(start)
	return repos, stats, nil
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

// warmRefreshStats is phase-level refresh timing for Debug logs.
type warmRefreshStats struct {
	budget    time.Duration
	eligible  int
	refreshed int
	duration  time.Duration
}

// warmBudgetRefresh rewalks oldest eligible direct-child units under absRoot
// until WarmRefreshBudget wall time is exhausted, merging newly found repos
// into existing. Negative budget means no refresh work. Zero budget uses the
// product default (1s). Cold paths never call this.
//
// Budget is enforced mid-unit: each unit rewalk uses a child context whose
// deadline is the remaining budget (parent ctx is not cancelled). On child
// deadline/cancel while the parent is still live, the unit stop is soft —
// partial discoveries merge into existing and remaining units are skipped.
//
// Units without a parseable mirror refreshed_at are not eligible (so brand-new
// root-level dirs planted after cold seed stay soft-incomplete under default
// YoungAge, matching P3 warm serve-only behavior).
func warmBudgetRefresh(ctx context.Context, absRoot string, opts Options, cacheRoot string, ignore ignoreConfig, existing []Repo, onRepo func(Repo) error) ([]Repo, warmRefreshStats, error) {
	start := time.Now()
	var stats warmRefreshStats

	budget, ok := resolveWarmRefreshBudget(opts.WarmRefreshBudget)
	if !ok {
		stats.budget = 0
		stats.duration = time.Since(start)
		return existing, stats, nil
	}
	stats.budget = budget
	youngAge := resolveYoungAge(opts.YoungAge)
	now := resolveNow(opts)

	units, err := listRefreshUnits(cacheRoot, absRoot)
	if err != nil {
		stats.duration = time.Since(start)
		return existing, stats, err
	}
	// No direct-child units: optionally rewalk root once (design fallback).
	if len(units) == 0 {
		units = []string{absRoot}
	}

	candidates := eligibleRefreshUnits(cacheRoot, units, now, youngAge)
	stats.eligible = len(candidates)
	if len(candidates) == 0 {
		stats.duration = time.Since(start)
		return existing, stats, nil
	}

	seen := make(map[string]struct{}, len(existing))
	for _, r := range existing {
		seen[r.Path] = struct{}{}
	}

	mergeFound := func(found []Repo) error {
		for _, repo := range found {
			if _, dup := seen[repo.Path]; dup {
				continue
			}
			seen[repo.Path] = struct{}{}
			if onRepo != nil {
				if err := onRepo(repo); err != nil {
					return err
				}
			}
			existing = append(existing, repo)
		}
		return nil
	}

	budgetStart := time.Now()
	budgetDeadline := budgetStart.Add(budget)
	for _, u := range candidates {
		remaining := time.Until(budgetDeadline)
		if remaining <= 0 {
			break
		}
		select {
		case <-ctx.Done():
			stats.duration = time.Since(start)
			return existing, stats, ctx.Err()
		default:
		}

		// Mid-unit cap: child deadline = remaining budget wall (does not cancel
		// parent). Sub-ms remainings skip the child deadline so between-unit
		// gating still allows one small unit to finish under a tiny budget.
		walkCtx := ctx
		var unitCancel context.CancelFunc
		var unitCtx context.Context
		if remaining >= minBudgetForMidUnitCap {
			unitCtx, unitCancel = context.WithDeadline(ctx, budgetDeadline)
			walkCtx = unitCtx
		}
		found, walkErr := walkRoot(walkCtx, u.path, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr, nil, cacheRoot)
		if unitCancel != nil {
			unitCancel()
		}

		if walkErr != nil {
			// Parent cancel/deadline is a hard failure (SIGINT path).
			if err := ctx.Err(); err != nil {
				_ = mergeFound(found)
				stats.duration = time.Since(start)
				return existing, stats, err
			}
			// Child-only deadline/cancel: soft stop — keep warm serve + partial
			// merge, do not surface as RootError / Scan hard error.
			if unitCtx != nil && unitCtx.Err() != nil {
				if err := mergeFound(found); err != nil {
					stats.duration = time.Since(start)
					return existing, stats, err
				}
				stats.refreshed++
				break
			}
			stats.duration = time.Since(start)
			return existing, stats, walkErr
		}
		stats.refreshed++
		if err := mergeFound(found); err != nil {
			stats.duration = time.Since(start)
			return existing, stats, err
		}
	}
	stats.duration = time.Since(start)
	return existing, stats, nil
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
