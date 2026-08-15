# Scenario

History save uses v3.0 moves format after worktree and plain moves.

mvd --add repo → root only
mvd -w repo wt → worktree move (from_type=main, to_type=worktree)
mvd repo dst → plain move after worktree (from=root, to_type=main)

## Steps
- Create a git repo and add it to tracking.
- Create a worktree with `-w`.
- Plain-move the main repo to dst (final command writes history).

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

	req.Args = []string{"--add", repo}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("add: %s", resp.Output)
	}

	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", repo, wt}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree: %s", resp.Output)
	}

	dst := filepath.Join(req.WorkRoot, "dst")
	req.Args = []string{repo, dst}
	return nil
}
```