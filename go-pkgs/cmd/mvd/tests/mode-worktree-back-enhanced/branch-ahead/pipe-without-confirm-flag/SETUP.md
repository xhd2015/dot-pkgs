# Scenario

Piped stdin without --confirm-from-stdin must be rejected (no accidental merge).

mvd -w repo wt → [(repo), (wt w:wt)]
commit on wt → [feature branch ahead of main]
mvd --back wt (piped stdin, no flag) → error

## Steps
- Create a git repo, create a worktree, commit ahead of main.
- Run --back with piped stdin but WITHOUT --confirm-from-stdin.

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

	req.Args = []string{"--back", wtDir}
	req.StdinInput = "\n"
	return nil
}
```