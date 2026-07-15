# Scenario

**Feature**: budgeted warm refresh discovers a new repo under an aged unit

```
# cold seed unit container with known-repo
workspace/unit-a/known-repo  --cold seed--> mirror complete
stamp unit-a refreshed_at = now-2h  # older than YoungAge
plant workspace/unit-a/new-repo/.git  # uncached under aged unit
  -> Scan(NoCache=false, YoungAge=default/60s, WarmRefreshBudget=default/1s)
  -> Result includes known-repo AND new-repo
  -> mirror has is_repo for new-repo
```

## Steps

1. Create `unit-a/known-repo` under the scan root; cold-seed.
2. Stamp `unit-a` `refreshed_at` two hours in the past (eligible for refresh).
3. Plant `unit-a/new-repo` with fake `.git` after seed.
4. Leave `YoungAge` and `WarmRefreshBudget` at 0 so product defaults apply
   (60s / 1s) — enough budget and age without real sleeps.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	unitA := filepath.Join(root, "unit-a")
	known := filepath.Join(unitA, "known-repo")
	mkdirAll(t, known)
	fakeGitRepo(t, known)

	req.Roots = []string{root}
	req.NoCache = false
	// 0 → product defaults: YoungAge 60s, WarmRefreshBudget 1s
	req.YoungAge = 0
	req.WarmRefreshBudget = 0
	coldSeedScan(t, req.Roots, req.CacheRoot)

	unitPath := absPath(t, unitA)
	stampRefreshedAt(t, req.CacheRoot, unitPath, time.Now().Add(-2*time.Hour))

	newRepo := filepath.Join(unitA, "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)
	return nil
}
```
