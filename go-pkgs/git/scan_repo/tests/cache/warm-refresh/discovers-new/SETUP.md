# Scenario

**Feature**: budgeted warm refresh discovers a new repo under an aged unit

```
# cold seed unit container with known-repo
workspace/unit-a/known-repo  --cold seed--> home/repos.json
plant workspace/unit-a/nested/new-repo/.git  # nested: not sibling of known-repo
stamp unit-a ModTime = now-2h  # older than YoungAge (after plant)
  -> Scan(NoCache=false, YoungAge=default/60s, WarmRefreshBudget=default/1s)
  -> Result includes known-repo AND new-repo
  -> home/repos.json has new-repo
```

## Steps

1. Create `unit-a/known-repo` under the scan root; cold-seed.
2. Plant `unit-a/nested/new-repo` with fake `.git` after seed (needs rewalk).
3. Stamp `unit-a` ModTime two hours in the past (eligible for refresh).
4. Leave `YoungAge` and `WarmRefreshBudget` at 0 so product defaults apply.

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
	req.YoungAge = 0
	req.WarmRefreshBudget = 0
	coldSeedScan(t, req.Roots, req.CacheRoot)

	newRepo := filepath.Join(unitA, "nested", "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)

	// After plant so mtime is not refreshed by mkdir under unit.
	stampUnitModTime(t, unitA, time.Now().Add(-2*time.Hour))
	return nil
}
```
