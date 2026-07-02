# Scenario

**Feature**: non-diverged relations ignore tmp worktree — existing behavior

```
# relation is ahead/same/ancestor, worktree dirty → still errors "worktree not clean"
MergeBack -> IsClean -> error
```

## Steps

1. Create main repo with base commit.
2. Build the branch topology for each sub-scenario.
3. Make worktree dirty if requested.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		return err
	}
	initRepo(t, mainRepo, "")

	req.MainRepo = mainRepo
	return nil
}
```
