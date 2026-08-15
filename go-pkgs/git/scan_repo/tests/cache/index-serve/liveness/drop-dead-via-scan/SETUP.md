# Scenario

**Feature**: after cold indexes two mains, remove one `.git`; next Scan omits it

```
# live-repo + doomed-repo cold-seeded into mirror + index
then os.RemoveAll(doomed-repo/.git)  # path may remain; .git gone
  -> Scan(NoCache=false)
  -> Result includes live-repo only
  -> doomed-repo never listed
```

## Steps

1. Create `live-repo/` and `doomed-repo/` with fake `.git`.
2. Cold-seed; stash `DeadPath` = abs(doomed-repo), `KnownPath` = abs(live-repo).
3. Remove only `doomed-repo/.git` (directory gone; path may still exist).

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	live := filepath.Join(root, "live-repo")
	doomed := filepath.Join(root, "doomed-repo")
	for _, dir := range []string{live, doomed} {
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}

	req.Roots = []string{root}
	req.NoCache = false
	req.Refresh = false
	coldSeedScan(t, req.Roots, req.CacheRoot)

	req.KnownPath = absPath(t, live)
	req.DeadPath = absPath(t, doomed)

	// Kill git marker only — proves liveness on .git, not full path delete.
	if err := os.RemoveAll(filepath.Join(doomed, ".git")); err != nil {
		t.Fatalf("remove doomed .git: %v", err)
	}
	return nil
}
```
