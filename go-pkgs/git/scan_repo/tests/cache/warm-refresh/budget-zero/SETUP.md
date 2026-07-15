# Scenario

**Feature**: negative WarmRefreshBudget performs no unit rewalk (pure P3 serve)

```
# aged unit would be eligible, but budget forbids work
workspace/unit-a/known-repo --cold seed--> stamp unit-a now-2h
plant unit-a/new-repo
  -> Scan(YoungAge=1s, WarmRefreshBudget=-1)  # negative = zero refresh work
  -> Result includes known-repo only
  -> new-repo omitted (budget gate)
```

## Steps

1. Create `unit-a/known-repo`; cold-seed; stamp unit aged.
2. Plant `unit-a/new-repo`.
3. Set `WarmRefreshBudget=-1` (no refresh work); `YoungAge=1s` so age alone
   would allow refresh.

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
	req.YoungAge = time.Second
	// Negative budget: no refresh work (0 would mean product default 1s).
	req.WarmRefreshBudget = -1
	coldSeedScan(t, req.Roots, req.CacheRoot)

	stampRefreshedAt(t, req.CacheRoot, absPath(t, unitA), time.Now().Add(-2*time.Hour))

	newRepo := filepath.Join(unitA, "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)
	return nil
}
```
