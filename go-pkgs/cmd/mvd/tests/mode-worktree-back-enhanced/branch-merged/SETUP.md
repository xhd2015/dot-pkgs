# Scenario

Branch already merged into main. Back succeeds without prompt or rebase.

mvd -w repo wt → [(repo), (wt w:wt)]
commit on wt, merge to main → [branch merged]
mvd --back wt → success → remove wt+branch

## Steps
- Create a worktree at work/feature from main repo.
- Commit work on the feature branch.
- Merge feature into main.
- Run --back.

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

	writeFile(t, filepath.Join(wtDir, "feature-work"), "work")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work")

	runGit(t, mainRepo, "merge", "feature")

	req.Args = []string{"--back", wtDir}
	return nil
}
```
