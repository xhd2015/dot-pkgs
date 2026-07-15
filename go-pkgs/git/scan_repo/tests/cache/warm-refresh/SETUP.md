# Scenario

**Feature**: warm Scan spends a bounded budget refreshing oldest units (P4)

```
# budgeted refresh on warm-eligible root
NoCache=false + CacheRoot + root scan_complete
  -> warm serve (P3 liveness) then rewalk refresh units
  -> unit = direct child of scan root
  -> candidates: now - refreshed_at >= YoungAge (default 60s when 0)
  -> oldest refreshed_at first; skip young
  -> until WarmRefreshBudget exhausted (0 → default 1s; negative → no work)
  -> rewalk merges new repos into Result + updates mirror

# test control (no real 1s sleeps)
stamp unit refreshed_at via SaveCacheEntry
  + YoungAge / WarmRefreshBudget on Options
```

## Preconditions

- `CacheOp` empty so `Run` dispatches to `Scan`.
- `CacheRoot` is the temp dir from parent `cache/SETUP.md` (shared seed + warm Scan).
- Leaves cold-seed, then stamp unit ages / plant repos / set budget options.
- Fake `.git` fixtures; no enrichment.
- Refresh unit layout: repos live **under** a direct-child container (e.g.
  `root/unit-a/known-repo`), not as root-level siblings alone.

## Steps

1. Clear `CacheOp`; default `NoCache=false` (warm-eligible).
2. Provide `fakeGitRepo`, `coldSeedScan`, and `stampRefreshedAt` for descendants.

```go
import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, req *Request) error {
	req.CacheOp = ""
	req.NoCache = false
	req.ListRemotes = false
	req.ListWorktrees = false
	return nil
}

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

// coldSeedScan runs a full cold Scan that populates the mirror under CacheRoot
// (P2). Leaves mutate ages / FS afterward; Run exercises warm + budget refresh.
func coldSeedScan(t *testing.T, roots []string, cacheRoot string) {
	t.Helper()
	if cacheRoot == "" {
		t.Fatal("coldSeedScan: empty cacheRoot")
	}
	if len(roots) == 0 {
		t.Fatal("coldSeedScan: empty roots")
	}
	_, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     roots,
		CacheRoot: cacheRoot,
		NoCache:   false,
	})
	if err != nil {
		t.Fatalf("cold seed Scan: %v", err)
	}
	rootPath := absPath(t, roots[0])
	entry, ok, loadErr := scan_repo.LoadCacheEntry(cacheRoot, rootPath)
	if loadErr != nil {
		t.Fatalf("cold seed LoadCacheEntry(root): %v", loadErr)
	}
	if !ok {
		t.Fatalf("cold seed: expected mirror entry for scan root %s", rootPath)
	}
	if !entry.ScanComplete {
		t.Fatalf("cold seed: root ScanComplete=false, want true for warm eligibility")
	}
}

// stampRefreshedAt overwrites refreshed_at on an existing mirror entry so tests
// control unit age without sleeping.
func stampRefreshedAt(t *testing.T, cacheRoot, realPath string, at time.Time) {
	t.Helper()
	entry, ok, err := scan_repo.LoadCacheEntry(cacheRoot, realPath)
	if err != nil {
		t.Fatalf("stampRefreshedAt LoadCacheEntry(%s): %v", realPath, err)
	}
	if !ok {
		t.Fatalf("stampRefreshedAt: no mirror entry for %s", realPath)
	}
	entry.RefreshedAt = at.UTC().Format(time.RFC3339)
	if err := scan_repo.SaveCacheEntry(cacheRoot, realPath, entry); err != nil {
		t.Fatalf("stampRefreshedAt SaveCacheEntry(%s): %v", realPath, err)
	}
}
```
