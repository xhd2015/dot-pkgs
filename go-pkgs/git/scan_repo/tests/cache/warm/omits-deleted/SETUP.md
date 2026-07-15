# Scenario

**Feature**: warm Scan never emits deleted repos; clears stale `is_repo` cache marks

```
# liveness after cold seed
workspace/still-here + workspace/gone-repo  --cold seed--> both is_repo in mirror
then remove gone-repo/ entirely (no path, no .git)
  -> Scan(NoCache=false, same CacheRoot)
  -> Result includes still-here only
  -> gone-repo not listed
  -> cache for gone-repo: entry absent OR is_repo=false
```

## Steps

1. Create workspace with two main repos: `still-here/` and `gone-repo/`.
2. Cold-seed Scan into `req.CacheRoot`.
3. Remove `gone-repo/` entirely (directory gone — stale mirror `is_repo` remains until liveness).
4. Set `req.Roots`; `NoCache=false`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	stillHere := filepath.Join(root, "still-here")
	goneRepo := filepath.Join(root, "gone-repo")
	for _, dir := range []string{stillHere, goneRepo} {
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}

	req.Roots = []string{root}
	req.NoCache = false
	coldSeedScan(t, req.Roots, req.CacheRoot)

	// Stash absolute path for Assert (mirror key after real dir is gone).
	req.RealPath = absPath(t, goneRepo)

	if err := os.RemoveAll(goneRepo); err != nil {
		t.Fatalf("remove gone-repo: %v", err)
	}
	return nil
}
```
