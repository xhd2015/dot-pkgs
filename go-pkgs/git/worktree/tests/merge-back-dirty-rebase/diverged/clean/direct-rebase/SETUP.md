# Scenario

**Feature**: diverged + clean + no-rm → direct rebase in source (existing behavior, no tmp)

```
# clean worktree, !Remove → rebase directly in source, no tmp worktree created
clean feat -> MergeBack(!Remove) -> rebase in source -> merge -> source rebased
```

## Steps

1. Source worktree is clean (already set by parent).
2. Call MergeBack.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	// set WRK_HOME so we can assert no tmp was created
	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		return err
	}
	t.Setenv("WRK_HOME", wrkHome)
	return nil
}
```
