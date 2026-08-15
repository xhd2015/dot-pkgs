# Scenario

**Feature**: warm Scan returns previously indexed live repos and omits brand-new uncached ones that are not siblings of indexed repos

```
# prove warm (not full re-walk); sibling probe does not cover other top-level units
workspace/unit-a/known-repo  --cold seed--> home/repos.json
then plant workspace/unit-elsewhere/brand-new-repo/.git
  # unit-elsewhere is never a parent of any indexed repo → not sibling-probed
  -> Scan(NoCache=false, same CacheRoot)
  -> Result includes known-repo
  -> Result omits brand-new-repo  # soft incompleteness; no full walk / no P4 rewalk of unit-elsewhere
```

## Steps

1. Create workspace with one main repo `unit-a/known-repo/`.
2. Cold-seed Scan into `req.CacheRoot` (parent temp).
3. Plant `unit-elsewhere/brand-new-repo/` with fake `.git` after the seed.
4. Set `req.Roots`; keep `NoCache=false` so Run is warm-eligible.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	known := filepath.Join(root, "unit-a", "known-repo")
	mkdirAll(t, known)
	fakeGitRepo(t, known)

	req.Roots = []string{root}
	req.NoCache = false
	// Soft-omit proof: disable budgeted unit refresh so brand-new under another unit stays omitted.
	req.WarmRefreshBudget = -1
	coldSeedScan(t, req.Roots, req.CacheRoot)

	brandNew := filepath.Join(root, "unit-elsewhere", "brand-new-repo")
	mkdirAll(t, brandNew)
	fakeGitRepo(t, brandNew)
	return nil
}
```
