# Scenario

**Feature**: diverged + dirty + no-rm → tmp worktree succeeds

```
# dirty worktree, !Remove → create tmp, rebase there, merge, force-update branch, cleanup
dirty feat -> MergeBack(!Remove) -> tmp worktree -> rebase -> merge -> cleanup -> branch force-updated
```

## Steps

1. Set `WRK_HOME` to a temp dir so tmp worktrees are confined.
2. Call MergeBack.
3. Assert tmp worktree was cleaned up, source preserved, branch updated.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		return err
	}
	t.Setenv("WRK_HOME", wrkHome)
	return nil
}
```
