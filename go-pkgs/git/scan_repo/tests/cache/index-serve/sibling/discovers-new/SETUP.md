# Scenario

**Feature**: warm Scan finds uncached sibling `B` next to cold-indexed `A`

```
# fixture layout (single parent under scan root)
workspace/A/.git  --cold seed--> indexed + warm-eligible
then mkdir workspace/B + fake .git  (sibling of A; never in cold index)
  -> Scan(NoCache=false, Refresh=false)
  -> Result includes abs(A) and abs(B), both main
```

## Steps

1. Create workspace with only `A/`; cold-seed.
2. Plant sibling `B/` with fake `.git` after seed.
3. Stash `KnownPath=A`, `SiblingPath=B`; `Refresh=false`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	a := filepath.Join(root, "A")
	mkdirAll(t, a)
	fakeGitRepo(t, a)

	req.Roots = []string{root}
	req.NoCache = false
	req.Refresh = false
	coldSeedScan(t, req.Roots, req.CacheRoot)

	b := filepath.Join(root, "B")
	mkdirAll(t, b)
	fakeGitRepo(t, b)

	req.KnownPath = absPath(t, a)
	req.SiblingPath = absPath(t, b)
	return nil
}
```
