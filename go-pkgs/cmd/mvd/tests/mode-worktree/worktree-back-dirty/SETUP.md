# Scenario

Error when worktree has uncommitted changes.

mvd -w repo wt → [(repo), (wt w:wt)]
touch wt/dirty → [dirty worktree]
mvd --back wt → error → uncommitted changes

## Steps
- Create a worktree at work/feature.
- Write an uncommitted file in the worktree.
- Try --back.

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

	writeFile(t, filepath.Join(wtDir, "dirty-file"), "uncommitted")

	req.Args = []string{"--back", wtDir}
	return nil
}
```
