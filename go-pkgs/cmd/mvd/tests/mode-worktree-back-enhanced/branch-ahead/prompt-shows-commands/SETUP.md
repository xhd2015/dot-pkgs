# Scenario

**Feature**: ahead-branch confirmation prompt lists concrete git commands before [Y/n]

```
# branch ahead of main: --confirm shows FormatPlanPrompt; user declines
mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature branch ahead of main]
mvd --back --confirm --confirm-from-stdin wt → FormatPlanPrompt → 'n' → abort
```

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch (branch is now ahead of main HEAD).
- Run `--back --confirm --confirm-from-stdin` with `n` input (decline).
- Assert FormatPlanPrompt content (requires `--confirm` so prompts are not auto-yes'd).

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

	writeFile(t, filepath.Join(wtDir, "feature-work"), "ahead of main")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work ahead")

	req.Args = []string{"--back", "--confirm", "--confirm-from-stdin", wtDir}
	req.StdinInput = "n\n"
	return nil
}
```
