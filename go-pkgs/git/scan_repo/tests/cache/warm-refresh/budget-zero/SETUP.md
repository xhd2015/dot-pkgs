# Scenario

**Feature**: negative WarmRefreshBudget performs no unit rewalk (pure P3 serve + sibling probe)

```
# aged unit would be eligible, but budget forbids rewalk work
workspace/unit-a/known-repo --cold seed--> stamp unit-a now-2h
plant unit-a/nested/new-repo  # nested under unit-a; NOT a sibling of known-repo
  -> Scan(YoungAge=1s, WarmRefreshBudget=-1)  # negative = zero refresh work
  -> Result includes known-repo only
  -> new-repo omitted (budget gate; sibling probe does not see nested path)
```

## Steps

1. Create `unit-a/known-repo`; cold-seed; stamp unit aged.
2. Plant `unit-a/nested/new-repo` (requires unit rewalk; not a direct sibling of
   the indexed repo, so sibling probe alone cannot find it).
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

	stampUnitModTime(t, unitA, time.Now().Add(-2*time.Hour))

	// Nested (not sibling of known-repo under unit-a): only a unit rewalk finds it.
	newRepo := filepath.Join(unitA, "nested", "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)
	return nil
}
```
