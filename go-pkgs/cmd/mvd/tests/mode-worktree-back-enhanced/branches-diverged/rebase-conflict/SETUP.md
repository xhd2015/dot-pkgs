# Scenario

**Feature**: diverged conflict under default auto-yes aborts rebase and errors

```
# diverged with conflicting edits; auto-yes proceeds into rebase → CONFLICT
mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit conflicting change on wt
commit different conflicting change on main → [diverged + conflict]
mvd --back wt → rebase → CONFLICT → abort rebase, error
```

## Steps
- Create a git repo, create a worktree from it.
- Modify README.md in the worktree and commit.
- Modify README.md differently on main and commit.
- Run bare `--back` (default auto-yes). The rebase should conflict on README.md.

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

	// Modify README.md in worktree and commit.
	writeFile(t, filepath.Join(wtDir, "README.md"), "# feature change\n")
	runGit(t, wtDir, "add", "README.md")
	runGit(t, wtDir, "commit", "-m", "feature change to README")

	// Modify README.md differently on main and commit.
	writeFile(t, filepath.Join(mainRepo, "README.md"), "# main change\n")
	runGit(t, mainRepo, "add", "README.md")
	runGit(t, mainRepo, "commit", "-m", "main change to README")

	// Bare --back: default auto-yes proceeds into rebase which will conflict.
	req.Args = []string{"--back", wtDir}
	return nil
}
```
