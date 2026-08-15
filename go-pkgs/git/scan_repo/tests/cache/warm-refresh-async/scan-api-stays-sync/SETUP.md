# Scenario

**Feature**: classic `Scan` ignores Async mode and still merges refresh into Result

```
same fixture as discovers-new
  -> Scan(..., WarmRefreshMode=Async)  # forced sync inside Scan
  -> Result includes known + new
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
	// Parent Run may also Scan; Assert re-drives Scan with Async mode forced sync.
	req.WarmRefreshBudget = -1
	coldSeedScan(t, req.Roots, req.CacheRoot)

	newRepo := filepath.Join(unitA, "nested", "new-repo")
	mkdirAll(t, newRepo)
	fakeGitRepo(t, newRepo)
	stampUnitModTime(t, unitA, time.Now().Add(-2*time.Hour))
	return nil
}
```
