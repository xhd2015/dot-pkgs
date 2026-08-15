# Scenario

**Feature**: FindLocalMainByGitHub resolves main checkouts by github.com remote identity

```
FindLocalMainByGitHub(roots, owner, repo) -> first main match, skip worktrees
```

## Preconditions

- Real `git` on PATH (tests skip otherwise).
- Reuses enrich-remotes git helpers from parent SETUP chain.

## Steps

1. Leaf `Setup` builds fixture layout and sets `req.FindGitHubOwner` / `req.FindGitHubRepo`.

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
	return nil
}

func gitInitRepo(t *testing.T, dir string) {
	t.Helper()
	mkdirAll(t, dir)
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
}

func gitRemoteAdd(t *testing.T, dir, name, url string) {
	t.Helper()
	runGit(t, dir, "remote", "add", name, url)
}

func gitInitialCommit(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "README"), "init\n")
	runGit(t, dir, "add", "README")
	runGit(t, dir, "-c", "core.hooksPath=/dev/null", "commit", "-m", "init")
}
```