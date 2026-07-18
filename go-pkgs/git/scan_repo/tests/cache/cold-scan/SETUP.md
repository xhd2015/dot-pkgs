# Scenario

**Feature**: cold `Scan` with explicit `CacheRoot` seeds index + walk log (no dense mirror)

```
# cold Scan side effects (v2)
NoCache=false + CacheRoot set
  -> Scan full walk
  -> home/repos.json seed + home/walk.jsonl seal
  -> Result.Repos discovery unchanged
  -> no <CacheRoot>/mirror

# NoCache=true
  -> full walk; no index / walk / mirror under CacheRoot
```

## Preconditions

- `CacheOp` empty so `Run` dispatches to `Scan`.
- Asserts load durable index via `LoadRepoIndex` / walk files after Scan when writes enabled.
- Explicit temp `CacheRoot` from parent.

## Steps

1. Clear enrichment; leave Roots / NoCache to descendants.
2. Provide `fakeGitRepo` / `fakeGitWorktree` for fixtures.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.CacheOp = ""
	req.ListRemotes = false
	req.ListWorktrees = false
	return nil
}

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

func fakeGitWorktree(t *testing.T, mainDir, wtDir string) {
	t.Helper()
	fakeGitRepo(t, mainDir)
	mkdirAll(t, wtDir)
	// gitlink file pointing at main .git
	writeFile(t, filepath.Join(wtDir, ".git"), "gitdir: "+filepath.Join(mainDir, ".git")+"\n")
}

func assertNoMirrorDir(t *testing.T, cacheRoot string) {
	t.Helper()
	mirrorDir := filepath.Join(cacheRoot, "mirror")
	if st, err := os.Stat(mirrorDir); err == nil {
		t.Fatalf("mirror path exists at %s (mode=%v); dense mirror is retired", mirrorDir, st.Mode())
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat mirror: %v", err)
	}
}
```
