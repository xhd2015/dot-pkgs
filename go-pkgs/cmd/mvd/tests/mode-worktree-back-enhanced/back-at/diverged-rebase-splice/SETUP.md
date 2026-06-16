# Scenario

cmdWorktreeBackAt: branches diverged, rebase succeeds. The worktree entry is in the middle of the chain.
After rebase + ff merge, the worktree is removed and later entries are preserved.

Chain: [repo, mid, wt(feature), later]
mvd --back wt → rebase feature onto main(mid) → success → ff merge + remove wt → chain becomes [repo, mid, later]

## Steps
- Create a git repo at work/repo.
- Move repo to work/mid.
- Create a worktree from mid at work/wt.
- Move mid to work/later (creates a later entry after the worktree).
- Commit work on the feature branch (wt).
- Commit a different change on mid (diverging from feature).
- Run --back on the worktree path (wt).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "repo")
	mkdirAll(t, repo)
	initGitRepo(t, repo)

	// Step 1: plain move repo → mid
	mid := filepath.Join(req.WorkRoot, "mid")
	req.Args = []string{repo, mid}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("move repo→mid: %s", resp.Output)
	}

	// Step 2: create worktree from mid
	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", mid, wt}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("worktree add: %s", resp.Output)
	}

	// Step 3: plain move mid → later (creates entry after the worktree)
	later := filepath.Join(req.WorkRoot, "later")
	req.Args = []string{mid, later}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("move mid→later: %s", resp.Output)
	}

	// Step 4: Commit work on the feature branch (wt diverges).
	writeFile(t, filepath.Join(wt, "feature-work"), "feature content")
	runGit(t, wt, "add", "feature-work")
	runGit(t, wt, "commit", "-m", "feature work")

	// Step 5: Commit a different change on mid (now at 'later' path).
	writeFile(t, filepath.Join(later, "main-work"), "main content")
	runGit(t, later, "add", "main-work")
	runGit(t, later, "commit", "-m", "main work")

	// Step 6: --back on the worktree with TTY and Enter (confirm rebase).
	req.Args = []string{"--back", "--confirm-from-stdin", wt}
	req.StdinInput = "\n"
	return nil
}
```
