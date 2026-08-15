# Scenario

**Feature**: `Collect` discovers tags created in a temp git repo

```
git init + tags -> Collect(repoRoot) -> inventory matching tag list
```

## Preconditions

- Real `git` on PATH.

## Steps

1. Skip when `git` is unavailable.
2. Create temp repo with initial commit.
3. Create tags `v0.0.1`, `v0.0.2`, `sub/v0.0.1`, and `release-1.0`.
4. Set `req.RepoRoot` to the repo directory.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	root := t.TempDir()
	repoDir := filepath.Join(root, "proj")
	gitInitRepo(t, repoDir)
	gitInitialCommit(t, repoDir)
	gitCreateTag(t, repoDir, "v0.0.1")
	gitCreateTag(t, repoDir, "v0.0.2")
	gitCreateTag(t, repoDir, "sub/v0.0.1")
	gitCreateTag(t, repoDir, "release-1.0")
	req.RepoRoot = repoDir
	return nil
}
```