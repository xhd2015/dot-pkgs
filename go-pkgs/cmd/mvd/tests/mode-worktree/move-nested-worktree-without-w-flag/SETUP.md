# Scenario

Nested worktree .git file is updated after os.Rename move.

mvd --add wt → [(wt) w:wt, nested .git]
mvd wt dst → [(wt), (dst) .git updated]

## Steps
- Create a git repo at work/main.
- Manually add a worktree at work/readonly-master.
- From that worktree, create a nested worktree at work/pricing.
- Use mvd (without -w) to move the nested worktree to work/pricing-moved.

```go
import (
	"os/exec"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wt1Dir := filepath.Join(req.WorkRoot, "readonly-master")
	runGit(t, mainRepo, "worktree", "add", wt1Dir)

	wt2Dir := filepath.Join(req.WorkRoot, "pricing")
	cmd := exec.Command("git", "-C", wt1Dir, "worktree", "add", wt2Dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git worktree add from worktree1: %v\n%s", err, out)
	}

	wt2Dst := filepath.Join(req.WorkRoot, "pricing-moved")
	req.Args = []string{wt2Dir, wt2Dst}
	return nil
}
```
