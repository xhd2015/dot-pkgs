# Scenario

Same as above, but using basename resolution for the source.

mvd -w repo wt → [(repo), (wt w:wt)]
mvd repo dst → [(repo), (wt w:wt), (dst)]

## Steps
- Create a git repo at work/repo with one commit.
- Use `mvd -w repo wt` to create a worktree at work/wt.
- Use `mvd repo dst` via basename resolution (the bare name "repo", not the full path).

The plain move via basename should resolve to the main repo (not the worktree) and move the main repo directory to dst.

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

	// Step 1: create worktree
	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", repo, wt}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 2: plain move via basename "repo"
	dst := filepath.Join(req.WorkRoot, "dst")
	req.Args = []string{"repo", dst}
	return nil
}
```
