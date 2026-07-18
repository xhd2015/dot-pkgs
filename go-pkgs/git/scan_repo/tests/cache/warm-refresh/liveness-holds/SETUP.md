# Scenario

**Feature**: budgeted warm refresh still never emits deleted repos

```
# liveness + budgeted refresh together
workspace/unit-a/still-here + unit-a/gone-repo --cold seed
stamp unit-a aged; remove gone-repo/ entirely
  -> Scan(YoungAge=1s, WarmRefreshBudget=1s)
  -> Result includes still-here only
  -> gone-repo not listed; cache mark cleared
```

## Steps

1. Create `unit-a/still-here` and `unit-a/gone-repo`; cold-seed.
2. Stamp unit aged so refresh may rewalk; delete `gone-repo` entirely.
3. Stash deleted abs path in `req.RealPath` for Assert.
4. Set enough budget and small YoungAge so refresh is attempted.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	unitA := filepath.Join(root, "unit-a")
	stillHere := filepath.Join(unitA, "still-here")
	goneRepo := filepath.Join(unitA, "gone-repo")
	for _, dir := range []string{stillHere, goneRepo} {
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}

	req.Roots = []string{root}
	req.NoCache = false
	req.YoungAge = time.Second
	req.WarmRefreshBudget = time.Second
	coldSeedScan(t, req.Roots, req.CacheRoot)

	stampUnitModTime(t, unitA, time.Now().Add(-2*time.Hour))

	req.RealPath = absPath(t, goneRepo)
	if err := os.RemoveAll(goneRepo); err != nil {
		t.Fatalf("remove gone-repo: %v", err)
	}
	return nil
}
```
