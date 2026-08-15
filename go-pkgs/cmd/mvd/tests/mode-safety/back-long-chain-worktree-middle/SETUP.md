> `--back` on a long chain `[repo, wt(wt), A, B]` verifies that `cmdBack`
> skips the worktree entry when finding the previous non-worktree location.
> Not a bug — sanity check for the worktree-skipping logic.
# Scenario

Long chain; --back skips worktree entry when finding previous location.

mvd -w repo wt → [(repo), (wt w:wt)]
mvd repo A → [(repo), (wt w:wt), (A)]
mvd A B → [(repo), (wt w:wt), (A), (B)]
mvd --back B → [(repo), (wt w:wt), (A)]  (B removed)

## Steps
- Create a git repo at work/repo.
- Create worktree from the root: `mvd -w repo wt` → [repo, wt(wt)].
- Plain move from root: `mvd repo A` → [repo, wt(wt), A].
- Plain move: `mvd A B` → [repo, wt(wt), A, B].
- Then `mvd --back B` → should skip worktree wt and go back to A.

This tests the skip-worktree-in-prev logic in cmdBack. The worktree was created
from root (not from the immediate source A), which avoids a position-check bug
in resolveMoveSource that fails when absSrc is neither root nor last.

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

	// Step 1: create worktree from repo
	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", repo, wt}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 2: repo → A (plain move from root; absSrc=root passes position check)
	A := filepath.Join(req.WorkRoot, "A")
	req.Args = []string{repo, A}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move repo→A failed: %s", resp.Output)
	}

	// Step 3: A → B (plain move from last entry; absSrc=last passes position check)
	B := filepath.Join(req.WorkRoot, "B")
	req.Args = []string{A, B}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move A→B failed: %s", resp.Output)
	}

	// Step 4: --back B → should skip wrktree wt and go to A
	req.Args = []string{"--back", B}
	return nil
}
```
