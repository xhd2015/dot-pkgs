## Expected
- When the destination directory exists, mvd creates the worktree inside it (appending source basename).
- The worktree `.git` file exists at existing-dir/main/.git.
- History records the source repo and the new worktree path.
- Output contains "worktree created:".

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	srcRepo := filepath.Join(req.WorkRoot, "main")
	existingDir := filepath.Join(req.WorkRoot, "existing-dir")
	movedDir := filepath.Join(existingDir, "main")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "worktree created:")
	assertFileExists(t, filepath.Join(movedDir, ".git"))
	assertFileNotExists(t, filepath.Join(existingDir, ".git"))
	assertHistoryChain(t, req.ConfigHome, srcRepo, srcRepo, movedDir)
	assertHistoryWorktreeEntry(t, req.ConfigHome, srcRepo, 1, srcRepo, "main")
}
```
