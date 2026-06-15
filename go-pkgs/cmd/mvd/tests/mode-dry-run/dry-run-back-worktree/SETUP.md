## Preconditions
- Git must be available.

## Steps
- Create a git repo, create a worktree from it.
- Then dry-run `--back` on the worktree path.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	mainRepo := filepath.Join(req.WorkRoot, "main")
	mkdirAll(t, mainRepo)
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(req.WorkRoot, "feature")
	// First create the worktree normally
	req.Args = []string{"-w", mainRepo, wtDir}
	resp, err := runMvd(t, req)
	if err != nil { return err }
	if resp.ExitCode != 0 { t.Fatalf("worktree move: %s", resp.Output) }

	// The feature branch is at the same commit as master (same HEAD),
	// so merge-base --is-ancestor feature HEAD passes without a merge.

	// Now dry-run back on the worktree
	req.Args = []string{"--dry-run", "--back", wtDir}
	return nil
}
```
