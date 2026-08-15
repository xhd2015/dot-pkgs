# Scenario

**Feature**: tiny WarmRefreshBudget aborts mid-unit rewalk; warm-served seed repos still returned

```
# one aged unit with known seed + huge uncached pad tree
workspace/unit-a/known-repo  --cold seed--> stamp unit-a aged
plant unit-a/pad/d00000..dN (many empty non-repo dirs; full rewalk >> budget)
  -> Scan(YoungAge=1s, WarmRefreshBudget=100ms)
  -> Result still includes known-repo (warm serve / partial merge)
  -> wall time << unbounded walkRoot of unit-a (budget bounds mid-unit, not only between units)
  -> Scan hard-error is nil (budget cancel uses unit-scoped child ctx, not parent SIGINT ctx)
```

## Steps

1. Create `unit-a/known-repo`; cold-seed so root is warm-eligible and known is cached.
2. Stamp `unit-a` `refreshed_at` two hours ago so it is eligible under `YoungAge=1s`.
3. Plant a large pad of empty directories under `unit-a/pad/` **after** seed so the
   unit rewalk must visit thousands of uncached dirs (each may touch many dirs
   entries — intentionally slow if unbounded).
4. Set `WarmRefreshBudget=100ms` so a correct mid-unit cap finishes near budget;
   an unbounded single `walkRoot` of this unit takes many seconds.

```go
import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

// padDirCount is large enough that an unbounded unit rewalk with cache writes
// is multi-second on typical developer machines (bench ~6–12s at 1k–2k). 2k
// keeps a safe margin above the 2s Assert wall cap if a mid-unit budget is missing.
const padDirCount = 2000

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	unitA := filepath.Join(root, "unit-a")
	known := filepath.Join(unitA, "known-repo")
	mkdirAll(t, known)
	fakeGitRepo(t, known)

	req.Roots = []string{root}
	req.NoCache = false
	req.YoungAge = time.Second
	// Small positive budget: unit walk must honor remaining budget mid-walk
	// (child context deadline / ctx.Done checks), not only between units.
	req.WarmRefreshBudget = 100 * time.Millisecond
	coldSeedScan(t, req.Roots, req.CacheRoot)

	stampUnitModTime(t, unitA, time.Now().Add(-2*time.Hour))

	// Huge uncached subtree under the only eligible unit. Planted after seed so
	// warm serve still has known-repo while refresh rewalk pays the pad cost.
	padRoot := filepath.Join(unitA, "pad")
	for i := 0; i < padDirCount; i++ {
		mkdirAll(t, filepath.Join(padRoot, fmt.Sprintf("d%05d", i)))
	}
	return nil
}
```
