# Scenario

**Feature**: default auto-yes ff-merges ahead branch without confirm flags or stdin

```
# HEAD ancestor of worktree branch; bare --back auto-yes (no Proceed?)
mvd -w repo wt → [(repo), (wt w:wt, branch: feature)]
commit on wt → [feature branch ahead of main]
mvd --back wt → ff merge + remove wt + delete branch
```

## Steps
- Create a git repo, create a worktree from it.
- Commit work on the feature branch (branch is now ahead of HEAD).
- Run `--back` with no confirm flags and no stdin input.

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

	// Bare --back: default auto-yes (no --confirm, no --confirm-from-stdin).
	req.Args = []string{"--back", wtDir}
	return nil
}
```
