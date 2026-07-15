# Scenario

**Feature**: cold `Scan` with explicit `CacheRoot` populates (or skips) mirror entries

```
# cold walk + cache side effects
caller roots + CacheRoot + NoCache
  -> Scan (full live walk)
  -> Result.Repos (discovery)
  -> when NoCache=false: SaveCacheEntry for visited dirs under mirror/
  -> when NoCache=true: no mirror write
```

## Preconditions

- `CacheOp` remains empty so `Run` dispatches to `Scan` (not pure cache APIs).
- `CacheRoot` is a temp dir from parent `cache/SETUP.md`.
- Fake `.git` fixtures; no enrichment (`ListRemotes`/`ListWorktrees` false).
- Asserts load mirror entries via `LoadCacheEntry` after Scan.

## Steps

1. Clear `CacheOp` so Scan path runs.
2. Default `NoCache` false (write branch); `no-cache/` overrides.
3. Provide `fakeGitRepo` / `fakeGitWorktree` for fixtures (siblings under `scan/` are not inherited).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.CacheOp = ""
	req.NoCache = false
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
	wtName := filepath.Base(wtDir)
	wtGitDir := filepath.Join(mainDir, ".git", "worktrees", wtName)
	mkdirAll(t, wtGitDir)
	absWtGitDir := absPath(t, wtGitDir)
	writeFile(t, filepath.Join(wtDir, ".git"), "gitdir: "+absWtGitDir+"\n")
}
```
