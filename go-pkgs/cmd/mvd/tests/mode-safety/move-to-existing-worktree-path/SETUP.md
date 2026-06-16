> Plain move targeting a directory that IS an existing worktree.
> Not a bug — `cmdMove` sees it as an existing dir, joins the basename, and
> safely nests the moved dir inside the worktree.
# Scenario

Plain move targeting a path that IS an existing worktree. Joins basename safely.

mvd -w repo1 wt → [(repo1), (wt w:wt)]
mvd repo2 wt → [(repo1), (wt w:wt)]  +  [(repo2), (wt/repo2)]

## Steps
- Create a git repo at work/repo1.
- `mvd -w repo1 wt` to create a worktree at work/wt.
- Create a SECOND directory at work/repo2 (plain dir, not a git repo).
- `mvd repo2 wt` to plain move repo2 to wt.

Since wt/ exists as a directory (it's the worktree), cmdMove joins the basename
and the move target becomes wt/repo2. The plain directory repo2 is moved INSIDE
the worktree directory. The worktree itself (wt/) is NOT affected (its .git file
still references repo1). The new repo2 dir is just nested inside the worktree dir,
which is a confusing but valid filesystem state.

Note: repo2 is a plain directory, not a git repo, so no worktree .git update is
triggered. This tests that mvd doesn't crash or corrupt anything when a plain
move target overlaps with an existing worktree path.

```go
import (
	"fmt"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)

	// Create repo1 (git repo) and repo2 (plain dir)
	repo1 := filepath.Join(req.WorkRoot, "repo1")
	mkdirAll(t, repo1)
	initGitRepo(t, repo1)

	repo2 := filepath.Join(req.WorkRoot, "repo2")
	mkdirAll(t, repo2)
	writeFile(t, filepath.Join(repo2, "data.txt"), "repo2 content")

	// Step 1: create worktree from repo1
	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", repo1, wt}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 2: plain move repo2 → wt (wt exists as dir, so target becomes wt/repo2)
	req.Args = []string{repo2, wt}
	return nil
}
```
