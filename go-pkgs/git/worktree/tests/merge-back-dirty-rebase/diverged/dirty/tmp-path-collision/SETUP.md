# Scenario

**Feature**: diverged + dirty + no-rm → tmp worktree path collision → suffix

```
# tmp dir at expected path already exists → use suffix -1
existing tmp dir -> MergeBack -> create at -1 suffix -> succeed -> clean up suffix dir
```

## Steps

1. Pre-create the expected tmp worktree directory to trigger collision.
2. Call MergeBack with the test's work-root-local temporary-worktree parent.

```go
import (
	"os"
	"path/filepath"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	worktreesDir := filepath.Join(wrkHome, "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return err
	}

	// pre-create the expected tmp worktree dir (without suffix)
	// naming: <repoBasename>-<pathToken>-<date>-tmp-rebase
	date := time.Now().Format("2006-01-02")
	collisionDir := filepath.Join(worktreesDir, "main-feature-"+date+"-tmp-rebase")
	if err := os.MkdirAll(collisionDir, 0755); err != nil {
		return err
	}

	return nil
}
```
