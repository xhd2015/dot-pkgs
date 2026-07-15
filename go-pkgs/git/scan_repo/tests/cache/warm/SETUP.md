# Scenario

**Feature**: warm `Scan` serves from a complete mirror cache (P3) with liveness

```
# warm eligibility (prior cold write left scan_complete on root)
NoCache=false + CacheRoot set + root entry scan_complete=true
  -> Scan serves candidate repos from mirror is_repo entries
  -> liveness: real path/.git must still exist or omit + clear mark
  -> brand-new uncached repos may be missed (no full re-walk; no P4 budget)

# contrast: NoCache=true always full live walk (finds brand-new)
# contrast P5: Options.Refresh force-cold lives under nested force-refresh/ tree

# Setup pattern for leaves
cold seed Scan (populate mirror) -> mutate FS -> Run second Scan under test
```

## Preconditions

- `CacheOp` empty so `Run` dispatches to `Scan`.
- `CacheRoot` is the temp dir from parent `cache/SETUP.md` (shared across cold seed and warm Scan).
- Leaves call `coldSeedScan` after building the workspace so the root has a usable
  `scan_complete=true` entry before the second Scan under test.
- Fake `.git` fixtures; no enrichment.
- Warm product path is not required for seed — cold write (P2) already works.

## Steps

1. Clear `CacheOp`; default `NoCache=false` (warm-eligible).
2. Provide `fakeGitRepo` and `coldSeedScan` for descendants.

```go
import (
	"context"
	"path/filepath"
	"testing"

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
// (P2). Leaves mutate the workspace afterward; Run exercises the warm path.
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
	// Sanity: root must be usable for warm (scan_complete).
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
```
