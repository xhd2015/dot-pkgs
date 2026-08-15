# Scenario

**Feature**: warm Scan returns indexed main repos when filesystem is unchanged

```
# cold seed then warm
workspace/known-repo/.git
  --cold seed--> mirror scan_complete + home/repos.json
  -> Scan(NoCache=false, Refresh=false)  # Run under test
  -> Result includes known-repo (main)
  -> IndexOK; index entry path == known-repo abs
```

## Steps

1. Create workspace with one main `known-repo/`.
2. Cold-seed into `req.CacheRoot` (same CacheRoot as Run).
3. Stash `req.KnownPath`; do not plant extra repos (proves serve, not sibling).

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	known := filepath.Join(root, "known-repo")
	mkdirAll(t, known)
	fakeGitRepo(t, known)

	req.Roots = []string{root}
	req.NoCache = false
	req.Refresh = false
	coldSeedScan(t, req.Roots, req.CacheRoot)

	req.KnownPath = absPath(t, known)
	return nil
}
```
