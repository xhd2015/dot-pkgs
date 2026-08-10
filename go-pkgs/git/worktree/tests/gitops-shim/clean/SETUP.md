# Scenario

**Feature**: go-pkgs IsClean / IsCleanWrk porcelain surface

```
# grouping creates clean master repo; leaves may add untracked dirt
worktree.IsClean / IsCleanWrk on req.Dir
```

## Preconditions

- `git` available (root already checked).

## Steps

1. Create temp directory and initialize git repository on `master`.
2. Write `README.md`, commit with message `init`.
3. Set `req.Op=clean` and `req.Dir` to the repository path.
4. Leaf setups may add untracked files or leave the tree clean.

## Context

- go-pkgs API names: `IsClean(path) error`, `IsCleanWrk(path) (bool, error)`.
- Porcelain non-empty (including untracked) makes `IsClean` return error and
  `IsCleanWrk` return false (wrk taxonomy maps `??` → added).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	dir, err := os.MkdirTemp("", "gopkgs-gitops-shim-clean-*")
	if err != nil {
		return err
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	runGit(t, dir, "init")
	runGit(t, dir, "branch", "-M", "master")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("init\n"), 0644); err != nil {
		return err
	}
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "init")

	req.Op = "clean"
	req.Dir = dir
	return nil
}
```
