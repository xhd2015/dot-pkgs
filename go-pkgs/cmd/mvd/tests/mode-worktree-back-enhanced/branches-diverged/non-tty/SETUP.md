# Scenario

Branches have diverged (neither is ancestor of the other), but stdin is not a TTY.
The prompt cannot be shown → error.

mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature ahead]
commit on main → [main also ahead → diverged]
mvd --back wt (non-TTY) → error

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch.
- Commit a different change on main (creating divergence).
- Run --back WITHOUT PTY (stdin is not a TTY). No input is piped.

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

	// Run --back normally (no PTY, stdin is /dev/null → not a TTY).
	req.Args = []string{"--back", wtDir}
	// UseScript is false, StdinInput is empty → default non-TTY behavior
	return nil
}
```
