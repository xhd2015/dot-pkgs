# Scenario

Worktree creation using a registered alias.

mvd --add repo → [(repo)]
mvd --add-alias al repo → alias registered
mvd -w al wt → [(repo), (wt w:wt)]

## Steps
- Create a git repo under work/projects/myrepo.
- Add it to mvd history with --add.
- Register an alias "myalias" for the basename "myrepo".
- Change CWD to an empty directory to avoid local shadowing.
- Run mvd -w myalias to create a worktree.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	projectRoot := filepath.Join(req.WorkRoot, "projects")
	mainRepo := filepath.Join(projectRoot, "myrepo")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	req.Args = []string{"--add", mainRepo}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add: %s", resp.Output)
	}

	req.Args = []string{"--add-alias", "myalias", "myrepo"}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add-alias: %s", resp.Output)
	}

	cwd := filepath.Join(req.WorkRoot, "cwd")
	mkdirAll(t, cwd)
	if err := os.Chdir(cwd); err != nil {
		return err
	}

	wtDir := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"-w", "myalias", wtDir}
	return nil
}
```
