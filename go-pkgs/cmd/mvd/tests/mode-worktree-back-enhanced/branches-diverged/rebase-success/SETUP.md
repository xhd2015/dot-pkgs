# Scenario

Branches have diverged (neither is ancestor of the other). Rebase succeeds.
After rebase, fast-forward merge, remove worktree, delete branch.

mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature ahead]
commit on main → [main also ahead → diverged]
mvd --back wt → rebase feature onto main → success → ff merge + remove wt + delete branch

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch.
- Commit a different change on main (creating divergence).
- Run --back with TTY stdin and Enter (confirm the rebase prompt).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"-w", mainRepo, wtDir}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("worktree add: %s", resp.Output)
	}

	// Commit work on the feature branch.
	writeFile(t, filepath.Join(wtDir, "feature-work"), "feature content")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work")

	// Commit a different change on main (diverging).
	writeFile(t, filepath.Join(mainRepo, "main-work"), "main content")
	runGit(t, mainRepo, "add", "main-work")
	runGit(t, mainRepo, "commit", "-m", "main work")

	// Run --back with TTY and Enter (confirm rebase prompt).
	req.Args = []string{"--back", wtDir}
	req.StdinInput = "\n"
	req.UseScript = true
	return nil
}
```
