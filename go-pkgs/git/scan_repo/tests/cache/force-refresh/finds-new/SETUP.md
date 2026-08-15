# Scenario

**Feature**: `Options.Refresh=true` forces cold full walk so brand-new repos are found

```
# contrast to parent warm/serves-cached-omits-new
workspace/known-repo --cold seed--> complete cache
then plant brand-new-repo/.git
  -> Scan(Refresh=true, NoCache=false, CacheRoot still set)
  -> Result includes known-repo AND brand-new-repo (force cold full walk)
```

## Steps

1. Create workspace with `known-repo/`; cold-seed into `CacheRoot`.
2. Plant `brand-new-repo/` after seed.
3. Set `req.Refresh=true` so Run forces cold even though warm-eligible.

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
	coldSeedScan(t, req.Roots, req.CacheRoot)

	brandNew := filepath.Join(root, "brand-new-repo")
	mkdirAll(t, brandNew)
	fakeGitRepo(t, brandNew)

	req.Refresh = true
	return nil
}
```
