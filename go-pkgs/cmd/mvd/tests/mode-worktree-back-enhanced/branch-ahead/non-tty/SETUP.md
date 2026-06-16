# Scenario

HEAD is ancestor of the worktree branch, but stdin is not a TTY.
The prompt cannot be shown → error.

mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature branch ahead of main]
mvd --back wt (non-TTY) → error

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch (branch is now ahead of HEAD).
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

	// Commit work on the feature branch so it is ahead of main HEAD.
	writeFile(t, filepath.Join(wtDir, "feature-work"), "ahead of main")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work ahead")

	// Run --back normally (no PTY, stdin is /dev/null → not a TTY).
	req.Args = []string{"--back", wtDir}
	// UseScript is false, StdinInput is empty → default non-TTY behavior
	return nil
}
```
