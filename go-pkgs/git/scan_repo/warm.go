package scan_repo

import (
	"context"
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

// rootWarmEligible reports whether a scan root can be served warm from the
// durable repo index (entries under the root).
func rootWarmEligible(cacheRoot, absRoot string) bool {
	ok, _ := rootCacheMode(cacheRoot, absRoot, Options{})
	return ok
}

// rootCacheMode decides warm vs cold for a scan root and returns a greppable
// reason token for Debug logs (missing_root_entry, no_cache, refresh, ok).
// Warm eligibility is index-only: usable home/repos.json with entries under absRoot.
func rootCacheMode(cacheRoot, absRoot string, opts Options) (warm bool, reason string) {
	if opts.NoCache || cacheRoot == "" {
		return false, "no_cache"
	}
	if opts.Refresh {
		return false, "refresh"
	}
	idx, ok, err := LoadRepoIndex(cacheRoot, UniverseHome)
	if err != nil || !ok {
		return false, "missing_root_entry"
	}
	for _, e := range idx.Repos {
		if pathIsUnderRoot(absRoot, e.Path) {
			return true, "ok"
		}
	}
	return false, "missing_root_entry"
}

// warmServeStats is phase-level serve timing for Debug logs.
type warmServeStats struct {
	candidates int
	live       int
	duration   time.Duration
}

// warmServeRoot serves repos from the durable home universe index when present
// (LoadRepoIndex + ApplyLiveness + sibling ReadDir). Index-only; no mirror fallback.
func warmServeRoot(ctx context.Context, absRoot, cacheRoot string, onRepo func(Repo) error) ([]Repo, warmServeStats, error) {
	return warmServeRootOpts(ctx, absRoot, cacheRoot, Options{}, onRepo)
}

// warmServeRootOpts is warmServeRoot with Options for budget-aware sibling probe.
func warmServeRootOpts(ctx context.Context, absRoot, cacheRoot string, opts Options, onRepo func(Repo) error) ([]Repo, warmServeStats, error) {
	start := time.Now()
	var stats warmServeStats

	if cacheRoot == "" {
		stats.duration = time.Since(start)
		return nil, stats, nil
	}

	idx, ok, err := LoadRepoIndex(cacheRoot, UniverseHome)
	if err != nil {
		stats.duration = time.Since(start)
		return nil, stats, err
	}
	if !ok {
		stats.duration = time.Since(start)
		return nil, stats, nil
	}

	nUnder := 0
	for _, e := range idx.Repos {
		if pathIsUnderRoot(absRoot, e.Path) {
			nUnder++
		}
	}
	if nUnder == 0 {
		stats.duration = time.Since(start)
		return nil, stats, nil
	}
	return warmServeFromIndex(ctx, absRoot, cacheRoot, opts, idx, onRepo)
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
// Unit age is live directory ModTime (not mirror refreshed_at). Brand-new
// root-level dirs planted after cold seed stay young under default YoungAge.
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

	units, err := listRefreshUnits(absRoot)
	if err != nil {
		stats.duration = time.Since(start)
		return existing, stats, err
	}
	// No direct-child units: optionally rewalk root once (design fallback).
	if len(units) == 0 {
		units = []string{absRoot}
	}

	candidates := eligibleRefreshUnits(units, now, youngAge)
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
		// Warm unit refresh: live rewalk only; no mirror, no cold walk.jsonl visits.
		found, walkErr := walkRoot(walkCtx, u.path, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr, nil, cacheRoot, false)
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

	// Persist merged discoveries into durable index (no mirror).
	if cacheRoot != "" && stats.refreshed > 0 {
		if seedErr := seedHomeRepoIndex(cacheRoot, absRoot, existing); seedErr != nil {
			stats.duration = time.Since(start)
			return existing, stats, seedErr
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

// listRefreshUnits returns direct child directories of absRoot from the live filesystem.
func listRefreshUnits(absRoot string) ([]string, error) {
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
	return units, nil
}

type refreshUnit struct {
	path string
	at   time.Time
}

// eligibleRefreshUnits keeps units whose directory ModTime is at least youngAge
// old, sorted oldest-first (then path for stability).
func eligibleRefreshUnits(units []string, now time.Time, youngAge time.Duration) []refreshUnit {
	var out []refreshUnit
	for _, u := range units {
		info, err := os.Stat(u)
		if err != nil || !info.IsDir() {
			continue
		}
		at := info.ModTime()
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
