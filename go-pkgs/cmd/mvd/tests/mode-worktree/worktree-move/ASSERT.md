## Expected
- Exit code 0.
- Output contains "worktree created:" and "[branch: feature]".
- The worktree `.git` file exists.
- History records two locations with worktree metadata.

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "worktree created:")
	assertContains(t, resp.Output, "[branch: feature]")

	wtDir := filepath.Join(req.WorkRoot, "feature")
	assertFileExists(t, filepath.Join(wtDir, ".git"))

	mainRepo := filepath.Join(req.WorkRoot, "main")
	assertHistoryChain(t, req.ConfigHome, mainRepo, mainRepo, wtDir)
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```
