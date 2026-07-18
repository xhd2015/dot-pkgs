# Scenario

**Feature**: units younger than YoungAge are not rewalked even with budget

```
# unit just refreshed (within YoungAge)
workspace/unit-a/known-repo --cold seed--> unit-a refreshed_at ≈ now
plant unit-a/nested/new-repo/.git  # nested; not a sibling of known-repo
  -> Scan(YoungAge=1h, WarmRefreshBudget=1s)  # large budget, unit still young
  -> Result includes known-repo only
  -> new-repo omitted (unit not selected for refresh; sibling probe skips nested)
```

## Steps

1. Create `unit-a/known-repo`; cold-seed (unit `refreshed_at` is fresh).
2. Do **not** age the unit; plant `unit-a/nested/new-repo` (needs rewalk; not a
   direct sibling of the indexed repo).
3. Set `YoungAge=1h` so the unit is ineligible; `WarmRefreshBudget=1s` so lack of
   discovery is due to age gate, not budget.

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
	req.YoungAge = time.Hour
	req.WarmRefreshBudget = time.Second
	coldSeedScan(t, req.Roots, req.CacheRoot)

	// Fresh unit (no stamp): still within YoungAge.
	// Nested path is not discovered by sibling probe (parent of known is unit-a;
	// ReadDir only sees direct children with .git).
	newRepo := filepath.Join(unitA, "nested", "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)
	return nil
}
```
