# Scenario

Without -w flag, worktree dir is moved via os.Rename.

mvd --add wt → [(wt) w:wt exists]
mvd wt dst → [(wt), (dst) w:wt]

## Steps
- Create a git repo at work/main.
- Manually add a worktree at work/feature-wt.
- Use mvd (without -w) to move the worktree to work/feature-wt-moved.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(req.WorkRoot, "feature-wt")
	runGit(t, mainRepo, "worktree", "add", wtDir)

	wtDst := filepath.Join(req.WorkRoot, "feature-wt-moved")
	req.Args = []string{wtDir, wtDst}
	return nil
}
```
