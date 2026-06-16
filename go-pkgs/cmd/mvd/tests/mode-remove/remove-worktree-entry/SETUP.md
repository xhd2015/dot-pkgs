# Scenario

Remove a worktree entry from a chain without affecting the root or other worktrees.

mvd -w repo wt1 → [repo, wt1(w:repo)]
mvd -w repo wt2 → [repo, wt1(w:repo), wt2(w:repo)]
mvd --rm wt1 → [repo, wt2(w:repo)]  (wt1 removed, others intact)

## Steps
- Create a git repo at repo with one commit.
- Create two worktrees wt1 and wt2 using `-w`.
- Remove wt1 with `--rm`. This should remove only the wt1 entry from the chain.
- Root and wt2 should remain.

```go
import (
	"fmt"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)

	repo := filepath.Join(req.WorkRoot, "repo")
	mkdirAll(t, repo)
	initGitRepo(t, repo)

	// create first worktree
	wt1 := filepath.Join(req.WorkRoot, "wt1")
	req.Args = []string{"-w", repo, wt1}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree 1: %s", resp.Output)
	}

	// create second worktree
	wt2 := filepath.Join(req.WorkRoot, "wt2")
	req.Args = []string{"-w", repo, wt2}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree 2: %s", resp.Output)
	}

	// remove wt1 entry from chain
	req.Args = []string{"--rm", wt1}
	return nil
}
```
