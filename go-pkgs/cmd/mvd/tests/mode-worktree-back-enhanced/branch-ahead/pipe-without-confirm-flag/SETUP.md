# Scenario

**Feature**: piped stdin without confirm flags still auto-yes merges (CASE B)

```
# branch ahead; piped stdin without --confirm / --confirm-from-stdin → auto-yes
mvd -w repo wt → [(repo), (wt w:wt)]
commit on wt → [feature branch ahead of main]
mvd --back wt (piped stdin, no flags) → exit 0; ff-merge + remove
```

## Steps
- Create a git repo, create a worktree, commit ahead of main.
- Run `--back` with piped stdin but WITHOUT `--confirm` or `--confirm-from-stdin`.
- Default auto-yes must merge successfully (no accidental-merge guard requiring flags).

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

	// Piped stdin without confirm flags: default auto-yes.
	req.Args = []string{"--back", wtDir}
	req.StdinInput = "\n"
	return nil
}
```
