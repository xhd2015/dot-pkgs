# Scenario

HEAD is ancestor of the worktree branch. User accepts by pressing Enter (default Y).
Fast-forward merge, remove worktree, delete branch.

mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature branch ahead of main]
mvd --back wt → prompt [Y/n] → Enter (default Y) → ff merge + remove wt + delete branch

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch (branch is now ahead of HEAD).
- Run --back with TTY stdin and an empty line (Enter) as input.

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

	// Commit work on the feature branch so it is ahead of main HEAD.
	writeFile(t, filepath.Join(wtDir, "feature-work"), "ahead of main")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work ahead")

	// Run --back with PTY (TTY) and empty input (Enter = default Y).
	req.Args = []string{"--back", wtDir}
	req.StdinInput = "\n"
	req.UseScript = true
	return nil
}
```
