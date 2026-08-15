# Scenario

**Feature**: diverged branches auto-yes rebase then ff-merge without confirm flags

```
# neither ancestor; bare --back auto-yes → rebase + ff-merge + remove
mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature ahead]
commit on main → [main also ahead → diverged]
mvd --back wt → rebase feature onto main → success → ff merge + remove wt + delete branch
```

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch.
- Commit a different change on main (creating divergence).
- Run bare `--back` (default auto-yes; no confirm flags / stdin).

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

	// Bare --back: default auto-yes confirms the rebase plan.
	req.Args = []string{"--back", wtDir}
	return nil
}
```
