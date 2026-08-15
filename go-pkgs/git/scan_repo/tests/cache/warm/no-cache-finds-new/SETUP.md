# Scenario

**Feature**: `NoCache=true` bypasses warm and full-walks so brand-new repos are found

```
# contrast to warm soft incompleteness
workspace/known-repo --cold seed--> complete cache
then plant brand-new-repo/.git
  -> Scan(NoCache=true, CacheRoot still set)
  -> Result includes known-repo AND brand-new-repo (full live walk)
```

## Steps

1. Create workspace with `known-repo/`; cold-seed into `CacheRoot`.
2. Plant `brand-new-repo/` after seed.
3. Set `req.NoCache=true` so Run cannot take the warm path.

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
	// Seed while cache is enabled; then force NoCache for the Scan under test.
	coldSeedScan(t, req.Roots, req.CacheRoot)

	brandNew := filepath.Join(root, "brand-new-repo")
	mkdirAll(t, brandNew)
	fakeGitRepo(t, brandNew)

	req.NoCache = true
	return nil
}
```
