# Scenario

**Feature**: warm `Scan` serves from durable repo index with liveness (no dense mirror)

```
# warm eligibility (prior cold write left home/repos.json)
NoCache=false + CacheRoot set + usable home/repos.json under root
  -> Scan serves candidate repos from index
  -> liveness: real path/.git must still exist or omit + drop from index
  -> brand-new uncached repos outside sibling probe may be missed

# contrast: NoCache=true always full live walk (finds brand-new)
# contrast P5: Options.Refresh force-cold lives under nested force-refresh/ tree

# Setup pattern for leaves
cold seed Scan (populate home/repos.json) -> mutate FS -> Run second Scan under test
```

## Preconditions

- `CacheOp` empty so `Run` dispatches to `Scan`.
- `CacheRoot` is the temp dir from parent `cache/SETUP.md` (shared across cold seed and warm Scan).
- Leaves call `coldSeedScan` after building the workspace so home index exists before the second Scan.
- Fake `.git` fixtures; no enrichment.

## Steps

1. Clear `CacheOp`; default `NoCache=false` (warm-eligible via index).
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

// coldSeedScan runs a full cold Scan that seeds home/repos.json under CacheRoot.
// Leaves mutate the workspace afterward; Run exercises the warm path.
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
	// Sanity: home index must exist for warm (index-only warm eligibility).
	rootPath := absPath(t, roots[0])
	idx, ok, loadErr := scan_repo.LoadRepoIndex(cacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("cold seed LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatalf("cold seed: expected home/repos.json under %s", cacheRoot)
	}
	nUnder := 0
	for _, e := range idx.Repos {
		if e.Path == rootPath || filepath.HasPrefix(e.Path, rootPath+string(filepath.Separator)) {
			nUnder++
		}
	}
	if nUnder == 0 {
		t.Fatalf("cold seed: index has no entries under scan root %s; idx=%+v", rootPath, idx)
	}
}
```
