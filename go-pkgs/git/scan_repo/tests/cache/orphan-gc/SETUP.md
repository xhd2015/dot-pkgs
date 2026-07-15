# Scenario

**Feature**: orphan mirror GC (P7) — rewalked parents drop dead child mirror subtrees

```
# cold seed leaves mirror for every visited child
workspace/… + CacheRoot --cold seed--> entry.json for each child basename

# later: real child dir removed from disk; parent is rewalked
delete real gone/ (stale mirror entry remains)
  -> cold force Refresh OR budgeted unit rewalk of parent
  -> parent children list no longer includes gone
  -> mirror for gone: entry.json removed (not merely is_repo=false)

# contrast P3 warm liveness (warm/omits-deleted): may leave entry with is_repo=false
# P7 GC is stronger: remove dead path under mirror so cache does not grow forever
```

## Preconditions

- `CacheOp` empty so `Run` dispatches to `Scan`.
- `CacheRoot` is the temp dir from parent `cache/SETUP.md` (shared seed + rewalk Scan).
- Leaves cold-seed, then delete a real child, then force a parent rewalk
  (Refresh cold or aged unit refresh).
- Fake `.git` fixtures; no enrichment.
- `req.RealPath` stashes the absolute path of the deleted child for Assert
  (mirror key after the real dir is gone).

## Steps

1. Clear `CacheOp`; default `NoCache=false`.
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
	req.Refresh = false
	return nil
}

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

// coldSeedScan runs a full cold Scan that populates the mirror under CacheRoot.
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
		t.Fatalf("cold seed: root ScanComplete=false, want true")
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
