# Scenario

**Feature**: enrichment flags add git metadata to scan output

```
--list-remotes / --list-worktrees -> Scan enrichment -> formatted stdout
```

## Preconditions

- Real `git` on PATH (leaves skip otherwise).
- Default lines output (no `--json`).

## Steps

1. Provide git fixture helpers for init, remotes, and worktrees.

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
	// Enrich branch uses default lines output.
	var filtered []string
	for _, arg := range req.Args {
		if arg == "--json" {
			continue
		}
		filtered = append(filtered, arg)
	}
	req.Args = filtered
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

func gitWorktreeAdd(t *testing.T, mainDir, wtDir, branch string) {
	t.Helper()
	runGit(t, mainDir, "worktree", "add", "-b", branch, wtDir)
}
```