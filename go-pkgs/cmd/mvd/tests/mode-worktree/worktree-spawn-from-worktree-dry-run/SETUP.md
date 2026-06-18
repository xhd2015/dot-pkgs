# Scenario

--dry-run with -w when SRC is an existing linked git worktree.

main repo + wt1; mvd --dry-run -w wt1 wt2 → prints 'would create worktree'

## Preconditions
- Git must be available.

## Steps
- Create a git repo at work/main.
- Add a linked worktree at work/wt1 via `git worktree add`.
- Run `mvd --dry-run -w wt1 wt2`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wt1 := filepath.Join(req.WorkRoot, "wt1")
	runGit(t, mainRepo, "worktree", "add", wt1)

	wt2 := filepath.Join(req.WorkRoot, "wt2")
	req.Args = []string{"--dry-run", "-w", wt1, wt2}
	return nil
}
```