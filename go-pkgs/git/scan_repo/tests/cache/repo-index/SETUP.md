# Scenario

**Feature**: durable per-universe `repos.json` under CacheRoot (load/save/liveness)

```
# repo index store (P1 pure I/O)
caller CacheRoot + universe + RepoIndex
  -> SaveRepoIndex -> <CacheRoot>/<universe>/repos.json (atomic)
  -> LoadRepoIndex(cacheRoot, universe) -> (index, ok, err)
  -> ApplyLiveness(index) drops entries without live .git
```

## Preconditions

- Nested root: own helpers; does not inherit parent tree `Setup`.
- Every leaf uses an explicit temp `CacheRoot` from `t.TempDir()` (never
  `$HOME/.cache/git-repo-scan`).
- Universes under test: `"home"` and `"root"` only.
- Schema v1 fields are asserted by leaves; product types live in package
  `scan_repo` (`RepoIndex`, `RepoIndexEntry`).
- Fake `.git` fixtures for liveness only; no Scan, no enrichment, no git CLI.

## Steps

1. Allocate a fresh temp `CacheRoot`.
2. Provide `mkdirAll` / `fakeGitRepo` for liveness leaves.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(abs)
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}

// fakeGitRepo plants a minimal main-repo .git directory (objects only).
func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

func Setup(t *testing.T, req *Request) error {
	req.CacheRoot = t.TempDir()
	return nil
}
```
