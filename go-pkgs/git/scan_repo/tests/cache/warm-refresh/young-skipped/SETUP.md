# Scenario

**Feature**: units younger than YoungAge are not rewalked even with budget

```
# unit just refreshed (within YoungAge)
workspace/unit-a/known-repo --cold seed--> unit-a refreshed_at ≈ now
plant unit-a/new-repo/.git
  -> Scan(YoungAge=1h, WarmRefreshBudget=1s)  # large budget, unit still young
  -> Result includes known-repo only
  -> new-repo omitted (unit not selected for refresh)
```

## Steps

1. Create `unit-a/known-repo`; cold-seed (unit `refreshed_at` is fresh).
2. Do **not** age the unit; plant `unit-a/new-repo`.
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
	newRepo := filepath.Join(unitA, "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)
	return nil
}
```
