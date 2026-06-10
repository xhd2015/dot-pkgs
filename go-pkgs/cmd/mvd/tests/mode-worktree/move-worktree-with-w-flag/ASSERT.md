## Expected
- Exit code 0.
- Output contains "worktree created:" and "[branch: feature-wt]".
- work/feature-wt exists with .git and README.md.
- History records the worktree chain.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "worktree created:")
	assertContains(t, resp.Output, "[branch: feature-wt]")

	wtDir := filepath.Join(req.WorkRoot, "feature-wt")
	assertFileExists(t, filepath.Join(wtDir, ".git"))
	assertFileExists(t, filepath.Join(wtDir, "README.md"))

	mainRepo := filepath.Join(req.WorkRoot, "main")
	assertHistoryChain(t, req.ConfigHome, mainRepo, mainRepo, wtDir)
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature-wt")
}
```
