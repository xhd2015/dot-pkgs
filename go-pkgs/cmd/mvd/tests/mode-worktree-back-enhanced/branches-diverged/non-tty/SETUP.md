# Scenario

**Feature**: non-TTY diverged `--back` auto-yes succeeds (no confirm flags)

```
# branches diverged; stdin not a TTY; default auto-yes rebase+merge
mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature ahead]
commit on main → [main also ahead → diverged]
mvd --back wt (non-TTY) → exit 0; rebase + ff-merge + remove
```

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch.
- Commit a different change on main (creating divergence).
- Run `--back` WITHOUT PTY and without confirm flags (stdin is not a TTY).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"-w", mainRepo, wtDir}
	resp, err := runMvd(t, d, req)
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

	// Bare --back on non-TTY: default auto-yes (no flags, no stdin).
	req.Args = []string{"--back", wtDir}
	// UseScript is false, StdinInput is empty → default non-TTY behavior
	return nil
}
```
