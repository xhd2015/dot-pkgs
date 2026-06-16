# Scenario

cmdWorktreeBackAt: HEAD is ancestor of the worktree branch. User confirms (Enter).
Fast-forward merge, remove worktree, splice chain preserving later entries.

Chain: [repo, mid, wt(feature), later]
mvd --back wt → prompt [Y/n] → Enter (default Y) → ff merge + remove wt → chain becomes [repo, mid, later]

## Steps
- Create a git repo at work/repo.
- Move repo to work/mid.
- Create a worktree from mid at work/wt.
- Move mid to work/later (creates a later entry after the worktree).
- Commit work on the feature branch (wt) — branch is ahead of main HEAD.
- Run --back on the worktree path (wt) with TTY input (Enter = default Y).

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

	// Step 4: Commit work on the feature branch (branch ahead of main HEAD).
	writeFile(t, filepath.Join(wt, "feature-work"), "ahead of main")
	runGit(t, wt, "add", "feature-work")
	runGit(t, wt, "commit", "-m", "feature work ahead")

	// Step 5: --back on the worktree with TTY input (Enter = default Y).
	req.Args = []string{"--back", "--confirm-from-stdin", wt}
	req.StdinInput = "\n"
	return nil
}
```
