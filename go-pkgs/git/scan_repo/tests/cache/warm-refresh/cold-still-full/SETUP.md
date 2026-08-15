# Scenario

**Feature**: cold Scan (empty cache) ignores refresh budget and fully discovers

```
# no prior cache/index — cold path
workspace/repo-a + workspace/repo-b  (never seeded)
  -> Scan(CacheRoot set, NoCache=false, YoungAge=1ns, WarmRefreshBudget=1µs)
  -> full unlimited walk finds both (budget options must not limit cold)
```

## Steps

1. Create two main repos under root; **do not** cold-seed (empty cache (no index)).
2. Set tiny budget / YoungAge so a mistaken budget application on cold would fail.
3. Run Scan — must find both repos.

```go
import (
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	repoA := filepath.Join(root, "repo-a")
	repoB := filepath.Join(root, "repo-b")
	for _, dir := range []string{repoA, repoB} {
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}

	req.Roots = []string{root}
	req.NoCache = false
	// Budget knobs must not throttle the cold full walk.
	req.YoungAge = time.Nanosecond
	req.WarmRefreshBudget = time.Microsecond
	// No coldSeedScan — empty cache forces cold path.
	return nil
}
```
