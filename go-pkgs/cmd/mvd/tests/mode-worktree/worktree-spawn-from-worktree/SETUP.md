# Scenario

Spawn a new worktree when SRC is an existing linked git worktree.

main repo + wt1; mvd -w wt1 wt2 → [(wt1), (wt2 w:wt2)]

## Preconditions
- Git must be available.

## Steps
- Create a git repo at work/main.
- Add a linked worktree at work/wt1 via `git worktree add`.
- Run `mvd -w wt1 wt2` to spawn wt2 from the worktree source.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wt1 := filepath.Join(req.WorkRoot, "wt1")
	runGit(t, mainRepo, "worktree", "add", wt1)

	wt2 := filepath.Join(req.WorkRoot, "wt2")
	req.Args = []string{"-w", wt1, wt2}
	return nil
}
```