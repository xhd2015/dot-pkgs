# Scenario

--dry-run with -w: prints intent, skips git worktree add.

mvd --dry-run -w main feature → prints 'would create worktree'

## Preconditions
- Git must be available.

## Steps
- Create a git repository at `main` under WorkRoot.
- Run `mvd --dry-run -w main feature` to dry-run a worktree creation.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)
	wtDir := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"--dry-run", "-w", mainRepo, wtDir}
	return nil
}
```
