# Scenario

**Feature**: non-TTY ahead `--back` auto-yes succeeds (no confirm flags)

```
# HEAD ancestor of worktree branch; stdin not a TTY; default auto-yes
mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature branch ahead of main]
mvd --back wt (non-TTY) → exit 0; ff-merge + remove
```

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch (branch is now ahead of HEAD).
- Run `--back` WITHOUT PTY and without confirm flags (stdin is not a TTY).

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

	// Bare --back on non-TTY: default auto-yes (no flags, no stdin).
	req.Args = []string{"--back", wtDir}
	// UseScript is false, StdinInput is empty → default non-TTY behavior
	return nil
}
```
