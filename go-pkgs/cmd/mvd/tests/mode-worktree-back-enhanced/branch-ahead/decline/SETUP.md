# Scenario

**Feature**: user declines ahead merge via `--confirm` + stdin `n`

```
# branch ahead; --confirm re-enables Y/n; --confirm-from-stdin for non-TTY pipe
mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature branch ahead of main]
mvd --back --confirm --confirm-from-stdin wt (n) → abort, no changes
```

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch (branch is now ahead of HEAD).
- Run `--back --confirm --confirm-from-stdin` with `n` input (decline).

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

	// Commit work on the feature branch so it is ahead of main HEAD.
	writeFile(t, filepath.Join(wtDir, "feature-work"), "ahead of main")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work ahead")

	// --confirm forces interactive plan prompt; --confirm-from-stdin for pipe.
	req.Args = []string{"--back", "--confirm", "--confirm-from-stdin", wtDir}
	req.StdinInput = "n\n"
	return nil
}
```
