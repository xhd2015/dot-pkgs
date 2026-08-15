# Scenario

**Feature**: Stop aborts min-budget wait; already-written index retained

```
warm async with long budget
  -> Stop immediately after ScanSession
  -> Join returns promptly (not full budget hang)
  -> no hard error; index file still valid (may or may not have new yet)
```

## Steps

```go
import (
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	unitA := filepath.Join(root, "unit-a")
	known := filepath.Join(unitA, "known-repo")
	mkdirAll(t, known)
	fakeGitRepo(t, known)

	req.Roots = []string{root}
	req.NoCache = false
	req.YoungAge = 0
	// Parent Run is sync Scan — disable refresh so it does not race Assert.
	req.WarmRefreshBudget = -1
	coldSeedScan(t, req.Roots, req.CacheRoot)

	newRepo := filepath.Join(unitA, "nested", "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)
	stampUnitModTime(t, unitA, time.Now().Add(-2*time.Hour))
	return nil
}
```
