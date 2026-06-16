# Scenario

Branches have diverged with conflicting changes to the same file.
Rebase hits a conflict → abort rebase, report error.

mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit conflicting change on wt
commit different conflicting change on main → [diverged + conflict]
mvd --back wt → rebase → CONFLICT → abort rebase, error

## Steps
- Create a git repo, create a worktree from it.
- Modify README.md in the worktree and commit.
- Modify README.md differently on main and commit.
- Run --back with TTY stdin and Enter (confirm rebase). The rebase should conflict on README.md.

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

	// Modify README.md in worktree and commit.
	writeFile(t, filepath.Join(wtDir, "README.md"), "# feature change\n")
	runGit(t, wtDir, "add", "README.md")
	runGit(t, wtDir, "commit", "-m", "feature change to README")

	// Modify README.md differently on main and commit.
	writeFile(t, filepath.Join(mainRepo, "README.md"), "# main change\n")
	runGit(t, mainRepo, "add", "README.md")
	runGit(t, mainRepo, "commit", "-m", "main change to README")

	// Run --back with TTY and Enter (confirm rebase). Rebase will conflict on README.md.
	req.Args = []string{"--back", "--confirm-from-stdin", wtDir}
	req.StdinInput = "\n"
	return nil
}
```
