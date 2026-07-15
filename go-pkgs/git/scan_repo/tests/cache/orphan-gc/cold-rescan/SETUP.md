# Scenario

**Feature**: cold force-rescan removes orphan mirror entry for a deleted root child

```
# cold seed with sibling mains keep/ + gone/
workspace/keep + workspace/gone --cold seed--> both is_repo in mirror
  root children includes "gone" and "keep"
then remove real gone/ entirely
  -> Scan(Refresh=true, NoCache=false, same CacheRoot)  # cold full rewalk
  -> Result: keep only
  -> mirror for gone: entry.json gone (Load ok=false)
  -> root children no longer lists "gone"; keep still cached
```

## Steps

1. Create workspace with two main repos: `keep/` and `gone/`.
2. Cold-seed Scan into `req.CacheRoot`.
3. Sanity: mirror entry for `gone/` exists after seed.
4. Remove `gone/` entirely; stash abs path in `req.RealPath`.
5. Set `req.Refresh=true` so Run force-colds and rewalks the parent root.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	keep := filepath.Join(root, "keep")
	gone := filepath.Join(root, "gone")
	for _, dir := range []string{keep, gone} {
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}

	req.Roots = []string{root}
	req.NoCache = false
	coldSeedScan(t, req.Roots, req.CacheRoot)

	goneAbs := absPath(t, gone)
	// Prove seed wrote the entry we expect GC to remove later.
	if _, ok, err := scan_repo.LoadCacheEntry(req.CacheRoot, goneAbs); err != nil {
		t.Fatalf("pre-delete LoadCacheEntry(gone): %v", err)
	} else if !ok {
		t.Fatal("pre-delete: expected mirror entry for gone after cold seed")
	}

	req.RealPath = goneAbs
	if err := os.RemoveAll(gone); err != nil {
		t.Fatalf("remove gone: %v", err)
	}

	req.Refresh = true
	return nil
}
```
