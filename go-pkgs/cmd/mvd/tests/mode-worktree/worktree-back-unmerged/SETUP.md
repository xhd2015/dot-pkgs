# Scenario

**Feature**: non-TTY unmerged (ahead) `--back` auto-yes succeeds

```
# HEAD ancestor of worktree branch; non-TTY bare --back → auto-yes ff-merge
mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature branch ahead of main]
mvd --back wt (non-TTY) → exit 0; ff-merge + remove
```

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch (branch is now ahead of HEAD).
- Run `--back` normally (no PTY, stdin is /dev/null → not a TTY; no confirm flags).

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

	writeFile(t, filepath.Join(wtDir, "feature-work"), "work")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work")

	req.Args = []string{"--back", wtDir}
	return nil
}
```
