# Scenario

**Feature**: `update-ref` moves a branch checked out in another worktree

```
worktree add -b wrk -> RunCLI -m … -> wrk and master both at new SHA
```

## Steps

1. Add a linked worktree on new branch `wrk` at the target SHA.
2. Append `-m` `move worktree tip`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.WorktreeDir = filepath.Join(t.TempDir(), "wrk")
	runGit(t, req.Dir, "worktree", "add", "-b", "wrk", req.WorktreeDir)
	req.Args = append(req.Args, "-m", "move worktree tip")
	return nil
}
```
