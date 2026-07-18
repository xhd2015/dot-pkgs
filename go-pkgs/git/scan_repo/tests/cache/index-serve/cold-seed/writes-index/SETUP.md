# Scenario

**Feature**: cold Scan writes `home/repos.json` listing discovered main repos

```
# two sibling mains under one root
workspace/alpha + workspace/zebra  (fake .git)
  -> Scan(CacheRoot, NoCache=false)  # cold; no prior cache
  -> Result has both mains
  -> LoadRepoIndex(home) ok; entries include both abs paths
  -> Index.Base = abs(scan root); Universe = home
```

## Steps

1. Create workspace with main repos `alpha/` and `zebra/`.
2. Set `req.Roots` to the workspace; keep `NoCache=false`.
3. Do **not** pre-call `coldSeedScan` — Run is the cold Scan that must seed the index.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	for _, name := range []string{"alpha", "zebra"} {
		dir := filepath.Join(root, name)
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}
	req.Roots = []string{root}
	req.NoCache = false
	req.Refresh = false
	return nil
}
```
