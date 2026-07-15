# Scenario

**Feature**: refresh selects oldest unit first; tiny budget covers only one unit

```
# two containers, both eligible, different ages
workspace/unit-a (refreshed_at now-2h) + unit-b (now-1h)
plant unit-a/a-new and unit-b/b-new after seed
  -> Scan(YoungAge=1s, WarmRefreshBudget=1µs)  # enough wall time for ~one unit
  -> Result includes a-known, b-known, a-new
  -> b-new omitted (unit-b not refreshed; oldest-first stopped after unit-a)
```

## Steps

1. Create `unit-a/a-known` and `unit-b/b-known`; cold-seed.
2. Stamp `unit-a` older (-2h) than `unit-b` (-1h); both past YoungAge=1s.
3. Plant `a-new` under unit-a and `b-new` under unit-b.
4. Set tiny `WarmRefreshBudget` so at most one unit rewalk completes (wall clock
   after first unit exceeds budget; no artificial 1s sleep).

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	unitA := filepath.Join(root, "unit-a")
	unitB := filepath.Join(root, "unit-b")
	aKnown := filepath.Join(unitA, "a-known")
	bKnown := filepath.Join(unitB, "b-known")
	for _, dir := range []string{aKnown, bKnown} {
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}

	req.Roots = []string{root}
	req.NoCache = false
	req.YoungAge = time.Second
	// Tiny positive budget: first unit rewalk spends real wall time >> 1µs,
	// so a second unit should not start. Prefer this over sleeping 1s.
	req.WarmRefreshBudget = time.Microsecond
	coldSeedScan(t, req.Roots, req.CacheRoot)

	now := time.Now()
	stampRefreshedAt(t, req.CacheRoot, absPath(t, unitA), now.Add(-2*time.Hour))
	stampRefreshedAt(t, req.CacheRoot, absPath(t, unitB), now.Add(-1*time.Hour))

	aNew := filepath.Join(unitA, "a-new")
	bNew := filepath.Join(unitB, "b-new")
	for _, dir := range []string{aNew, bNew} {
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}
	return nil
}
```
