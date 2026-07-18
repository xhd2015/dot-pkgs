# Scenario

**Feature**: warm Scan omits deleted repos and drops them from the durable index

```
workspace/still-here + workspace/gone-repo  --cold seed--> both in home/repos.json
delete gone-repo/
  -> Scan(NoCache=false)
  -> Result has still-here only
  -> index no longer lists gone-repo (liveness drop)
```

## Steps

1. Create `still-here` and `gone-repo` mains; cold-seed.
2. Remove `gone-repo/` entirely; stash abs path on `req.RealPath`.
3. Keep NoCache=false for warm.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	still := filepath.Join(root, "still-here")
	gone := filepath.Join(root, "gone-repo")
	mkdirAll(t, still)
	mkdirAll(t, gone)
	fakeGitRepo(t, still)
	fakeGitRepo(t, gone)

	req.Roots = []string{root}
	req.NoCache = false
	req.WarmRefreshBudget = -1
	coldSeedScan(t, req.Roots, req.CacheRoot)

	goneAbs := absPath(t, gone)
	req.RealPath = goneAbs
	if err := os.RemoveAll(gone); err != nil {
		t.Fatalf("remove gone-repo: %v", err)
	}
	return nil
}
```
