# Scenario

**Feature**: warm Scan spends a bounded budget refreshing oldest units (P4), without dense mirror

```
# budgeted refresh on warm-eligible root (index present)
NoCache=false + CacheRoot + home/repos.json
  -> warm serve (index + liveness) then rewalk refresh units
  -> unit = direct child of scan root (live FS)
  -> candidates: now - unit_age >= YoungAge (default 60s when 0)
  -> unit_age from live directory ModTime (mirror refreshed_at retired)
  -> oldest first; skip young
  -> until WarmRefreshBudget exhausted (0 → default 1s; negative → no work)
  -> rewalk merges new repos into Result + updates home/repos.json

# test control (no real 1s sleeps)
stamp unit ModTime via os.Chtimes
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

1. Clear `CacheOp`; default `NoCache=false` (warm-eligible via index).
2. Provide `fakeGitRepo`, `coldSeedScan`, and `stampUnitModTime` for descendants.

```go
import (
	"context"
	"os"
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

// coldSeedScan runs a full cold Scan that seeds home/repos.json under CacheRoot.
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
	idx, ok, loadErr := scan_repo.LoadRepoIndex(cacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("cold seed LoadRepoIndex: %v", loadErr)
	}
	if !ok || len(idx.Repos) == 0 {
		t.Fatalf("cold seed: expected non-empty home/repos.json under %s", cacheRoot)
	}
}

// stampUnitModTime sets the unit directory mtime so tests control age without sleeping.
// Product unit age uses live directory ModTime after dense mirror retirement.
func stampUnitModTime(t *testing.T, unitPath string, at time.Time) {
	t.Helper()
	if err := os.Chtimes(unitPath, at, at); err != nil {
		t.Fatalf("stampUnitModTime Chtimes(%s): %v", unitPath, err)
	}
}
```
