## Steps
- Create a worktree at work/feature.
- Commit work on the feature branch.
- Try --back without merging.

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
