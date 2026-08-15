# Scenario

**Feature**: async warm polish freezes Result at serve; Join updates durable index

```
cold seed unit-a/known-repo
plant unit-a/nested/new-repo
stamp unit-a old
  -> ScanSession(Async, WarmRefreshBudget=2s)
  -> before Join: Result has known only; index may still lack new
  -> Join: index has new-repo; Result unchanged (still known only)
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
	// Parent Run uses classic Scan (sync). Disable unit refresh there so it
	// does not merge new-repo into the index before Assert's ScanSession.
	req.WarmRefreshBudget = -1
	coldSeedScan(t, req.Roots, req.CacheRoot)

	newRepo := filepath.Join(unitA, "nested", "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)
	stampUnitModTime(t, unitA, time.Now().Add(-2*time.Hour))
	return nil
}
```
