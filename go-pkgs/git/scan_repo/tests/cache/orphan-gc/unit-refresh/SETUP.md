# Scenario

**Feature**: budgeted unit rewalk removes orphan mirror under the refreshed parent unit

```
# cold seed unit container with keep + gone
workspace/unit-a/keep + unit-a/gone --cold seed--> both is_repo in mirror
stamp unit-a refreshed_at = now-2h  # eligible for budgeted rewalk
then remove real unit-a/gone/ entirely
  -> Scan(NoCache=false, YoungAge=1s, WarmRefreshBudget=1s, Refresh=false)
  -> Result: unit-a/keep only
  -> mirror for gone: entry.json gone (Load ok=false)
  -> unit-a children no longer lists "gone"
```

## Steps

1. Create `unit-a/keep` and `unit-a/gone` mains; cold-seed.
2. Stamp `unit-a` aged so budgeted refresh rewalks the parent unit.
3. Sanity: mirror entry for `gone` exists after seed.
4. Remove `unit-a/gone/`; stash abs path in `req.RealPath`.
5. Set small YoungAge + enough WarmRefreshBudget; leave Refresh=false.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	unitA := filepath.Join(root, "unit-a")
	keep := filepath.Join(unitA, "keep")
	gone := filepath.Join(unitA, "gone")
	for _, dir := range []string{keep, gone} {
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}

	req.Roots = []string{root}
	req.NoCache = false
	req.Refresh = false
	req.YoungAge = time.Second
	req.WarmRefreshBudget = time.Second
	coldSeedScan(t, req.Roots, req.CacheRoot)

	stampRefreshedAt(t, req.CacheRoot, absPath(t, unitA), time.Now().Add(-2*time.Hour))

	goneAbs := absPath(t, gone)
	if _, ok, err := scan_repo.LoadCacheEntry(req.CacheRoot, goneAbs); err != nil {
		t.Fatalf("pre-delete LoadCacheEntry(gone): %v", err)
	} else if !ok {
		t.Fatal("pre-delete: expected mirror entry for gone after cold seed")
	}

	req.RealPath = goneAbs
	if err := os.RemoveAll(gone); err != nil {
		t.Fatalf("remove gone: %v", err)
	}
	return nil
}
```
