# Scenario

**Feature**: warm Scan returns previously cached live repos and omits brand-new uncached ones

```
# prove warm (not full re-walk)
workspace/known-repo  --cold seed--> mirror is_repo + root scan_complete
then plant workspace/brand-new-repo/.git (never written to cache)
  -> Scan(NoCache=false, same CacheRoot)
  -> Result includes known-repo
  -> Result omits brand-new-repo  # soft incompleteness; no P4 budget refresh
```

## Steps

1. Create workspace with one main repo `known-repo/`.
2. Cold-seed Scan into `req.CacheRoot` (parent temp).
3. Plant `brand-new-repo/` with fake `.git` after the seed (uncached).
4. Set `req.Roots`; keep `NoCache=false` so Run is warm-eligible.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	known := filepath.Join(root, "known-repo")
	mkdirAll(t, known)
	fakeGitRepo(t, known)

	req.Roots = []string{root}
	req.NoCache = false
	coldSeedScan(t, req.Roots, req.CacheRoot)

	// After cache is complete: plant an uncached repo that a full re-walk would find.
	brandNew := filepath.Join(root, "brand-new-repo")
	mkdirAll(t, brandNew)
	fakeGitRepo(t, brandNew)
	return nil
}
```
