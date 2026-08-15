# Scenario

Deep chain: multiple moves + worktree; plain move finds main repo.

mvd repo A → [(repo), (A)]
mvd A B → [(repo), (A), (B)]
mvd -w B wt → [(repo), (A), (B), (wt w:wt)]
mvd B dst → [(repo), (A), (B), (wt w:wt), (dst)]

## Steps
- Create a git repo at work/repo with one commit.
- Use `mvd repo A` to move the main repo to work/A.
- Use `mvd A B` to move again to work/B.
- Use `mvd -w B wt` to create a worktree from the main repo (now at B).
- Use `mvd repo dst` (plain move using original root basename) to move the main repo to dst.

The plain move should traverse the chain: root=[repo], skip worktree at wt, find the last non-worktree location B, and move B to dst.

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

	// Step 1: repo → A
	a := filepath.Join(req.WorkRoot, "A")
	req.Args = []string{repo, a}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move repo→A failed: %s", resp.Output)
	}

	// Step 2: A → B
	b := filepath.Join(req.WorkRoot, "B")
	req.Args = []string{a, b}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move A→B failed: %s", resp.Output)
	}

	// Step 3: create worktree from B
	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", b, wt}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 4: plain move using original basename "repo"
	dst := filepath.Join(req.WorkRoot, "dst")
	req.Args = []string{"repo", dst}
	return nil
}
```
