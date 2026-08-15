# Scenario

-w with $X/myrepo via lls env var expansion.

mvd -w $X/myrepo wt → [(projects/myrepo), (wt w:wt)]

## Steps
- Set up lls config with X env var.
- Create a git repo at projects/main.
- Move worktree with -w using $X/main to destination.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	skipIfNoGit(t)

	homeDir := filepath.Join(req.WorkRoot, ".lls-home")
	configDir := filepath.Join(homeDir, "Library", "Application Support", "lls")
	mkdirAll(t, configDir)
	writeFile(t, filepath.Join(configDir, "config.json"), `{"envs":["X"]}`)

	projectRoot := filepath.Join(req.WorkRoot, "projects")
	mainRepo := filepath.Join(projectRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"-w", "$X/main", wtDir}
	return nil
}
```
