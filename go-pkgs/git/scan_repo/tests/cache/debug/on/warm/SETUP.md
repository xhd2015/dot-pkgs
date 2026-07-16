# Scenario

**Feature**: second Scan after cold seed with Debug logs warm mode and serve timing

```
# warm debug (complete root cache)
workspace/my-repo --cold seed (Debug=false)--> root scan_complete
  -> Scan(Debug=true, NoCache=false, same CacheRoot)
  -> warm serve path
  -> stderr: scan: … mode=warm … serve candidates/live/duration
```

## Preconditions

- Cold seed must leave root `scan_complete=true` (warm-eligible).
- Seed Scan uses `Debug=false` so only the Run Scan's stderr is under test.
- `Debug=true` for Run via parent `on/`.

## Steps

1. Create workspace with one fake main repo `my-repo/`.
2. Cold-seed into `req.CacheRoot` (no debug).
3. Set `req.Roots`; keep `NoCache=false` so Run is warm-eligible.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "my-repo")
	mkdirAll(t, repoDir)
	fakeGitRepo(t, repoDir)

	req.Roots = []string{root}
	req.NoCache = false
	coldSeedScan(t, req.Roots, req.CacheRoot)
	// Debug remains true from on/; Run is the warm Scan under test.
	return nil
}
```
