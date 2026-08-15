# Scenario

**Feature**: `ListRemotes=true` enriches repos with remote metadata via git

```
Scan -> git remote + config URL -> Remotes[] on each row
```

## Preconditions

- Real `git` on PATH (tests skip otherwise).
- `ListRemotes=true`, `ListWorktrees=false`.

## Steps

1. Enable `ListRemotes`.
2. Provide git fixture helpers.

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
	req.ListRemotes = true
	req.ListWorktrees = false
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
	runGit(t, dir, "commit", "-m", "init")
}
```