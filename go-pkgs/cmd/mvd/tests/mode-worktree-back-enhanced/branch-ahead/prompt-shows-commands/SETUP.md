# Scenario

**Feature**: ahead-branch confirmation prompt lists concrete git commands before [Y/n]

```
# branch ahead of main: user declines after seeing planned commands
mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature branch ahead of main]
mvd --back wt --confirm-from-stdin → FormatPlanPrompt → 'n' → abort
```

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch (branch is now ahead of main HEAD).
- Run --back with `--confirm-from-stdin` and `n` input (decline).

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

	writeFile(t, filepath.Join(wtDir, "feature-work"), "ahead of main")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work ahead")

	req.Args = []string{"--back", "--confirm-from-stdin", wtDir}
	req.StdinInput = "n\n"
	return nil
}
```