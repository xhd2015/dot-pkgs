# Scenario

Worktree directory is already a worktree; -w triggers git worktree add.

mvd --add wt → [(wt) w:wt exists]
mvd -w wt new-wt → [(wt), (new-wt w:new-wt)]

## Steps
- Create a git repo at work/main.
- Use -w to create a worktree at work/feature-wt.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(req.WorkRoot, "feature-wt")
	req.Args = []string{"-w", mainRepo, wtDir}
	return nil
}
```
