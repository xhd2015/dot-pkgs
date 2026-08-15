# Scenario

**Feature**: opt-in async warm polish (`WarmRefreshAsync` + `ScanSession` / `Join`)

```
# durable-only async after warm serve
cold seed → plant new under aged unit
  -> ScanSession(WarmRefreshMode=Async)
  -> Result = serve snapshot only (no new-repo in Result yet)
  -> Join → home/repos.json gains new-repo

# min-budget / idle
Join with no remaining work → returns promptly (no forced sleep)

# Stop aborts min-budget wait; partial index kept
```

## Preconditions

- Parent `cache/SETUP.md` provides temp `CacheRoot`.
- Uses `fakeGitRepo` / `coldSeedScan` / `stampUnitModTime` from warm-refresh helpers
  (re-declared here for nested isolation if needed).

## Steps

1. Clear enrichment; default NoCache=false.
2. Provide helpers for descendants.

```go
import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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

func coldSeedScan(t *testing.T, roots []string, cacheRoot string) {
	t.Helper()
	if cacheRoot == "" {
		t.Fatal("coldSeedScan: empty cacheRoot")
	}
	_, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     roots,
		CacheRoot: cacheRoot,
		NoCache:   false,
	})
	if err != nil {
		t.Fatalf("cold seed Scan: %v", err)
	}
}

func stampUnitModTime(t *testing.T, unitPath string, at time.Time) {
	t.Helper()
	if err := os.Chtimes(unitPath, at, at); err != nil {
		t.Fatalf("stampUnitModTime Chtimes(%s): %v", unitPath, err)
	}
}
```
