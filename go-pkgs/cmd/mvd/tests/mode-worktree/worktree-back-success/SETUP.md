## Steps
- Create a worktree at work/feature.
- Commit work, merge to main, then --back.

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

	runGit(t, mainRepo, "merge", "feature")

	req.Args = []string{"--back", wtDir}
	return nil
}
```
