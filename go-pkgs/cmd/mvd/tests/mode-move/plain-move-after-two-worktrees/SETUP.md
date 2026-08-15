# Scenario

After two worktrees, plain move skips both to find the main repo.

mvd -w repo wt1 → [(repo), (wt1 w:wt1)]
mvd -w repo wt2 → [(repo), (wt1 w:wt1), (wt2 w:wt2)]
mvd repo dst → [(repo), (wt1 w:wt1), (wt2 w:wt2), (dst)]

## Steps
- Create a git repo at work/repo with one commit.
- Use `mvd -w repo wt1` to create first worktree.
- Use `mvd -w repo wt2` to create second worktree (both from the same main repo).
- Use `mvd repo dst` (plain move) to move the main repo to dst.

The plain move should skip both worktrees and find the main repo's location (repo), then move repo to dst. Both worktrees should have their .git files updated to reference dst.

```go
import (
	"fmt"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	skipIfNoGit(t)

	repo := filepath.Join(req.WorkRoot, "repo")
	mkdirAll(t, repo)
	initGitRepo(t, repo)

	// Step 1: create first worktree
	wt1 := filepath.Join(req.WorkRoot, "wt1")
	req.Args = []string{"-w", repo, wt1}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree wt1 create failed: %s", resp.Output)
	}

	// Step 2: create second worktree
	wt2 := filepath.Join(req.WorkRoot, "wt2")
	req.Args = []string{"-w", repo, wt2}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree wt2 create failed: %s", resp.Output)
	}

	// Step 3: plain move — should move repo (main repo), not wt1 or wt2
	dst := filepath.Join(req.WorkRoot, "dst")
	req.Args = []string{repo, dst}
	return nil
}
```
