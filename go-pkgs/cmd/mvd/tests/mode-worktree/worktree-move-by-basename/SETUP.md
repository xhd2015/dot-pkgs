# Scenario

Worktree creation using basename resolution.

mvd -w repo wt → [(repo), (wt w:wt)]

## Steps
- Create a project repo under work/projects/myrepo.
- Add it to mvd history.
- Change to an unrelated directory, then use -w with the repo basename.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	projectRoot := filepath.Join(req.WorkRoot, "projects")
	mainRepo := filepath.Join(projectRoot, "myrepo")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	req.Args = []string{"--add", mainRepo}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add: %s", resp.Output)
	}

	cwd := filepath.Join(req.WorkRoot, "cwd")
	mkdirAll(t, cwd)
	if err := os.Chdir(cwd); err != nil {
		return err
	}

	wtDir := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"-w", "myrepo", wtDir}
	return nil
}
```
