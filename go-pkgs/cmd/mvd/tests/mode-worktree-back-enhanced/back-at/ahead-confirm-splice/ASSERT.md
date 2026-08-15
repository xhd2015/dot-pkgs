## Expected
- Exit code 0.
- Output contains "worktree removed:".
- Worktree directory (work/wt) no longer exists.
- Main repo (work/later) has the feature-work file (ff merged).
- History chain: [repo, mid, later] — wt spliced out, later preserved.

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

	repo := filepath.Join(req.WorkRoot, "repo")
	mid := filepath.Join(req.WorkRoot, "mid")
	wt := filepath.Join(req.WorkRoot, "wt")
	later := filepath.Join(req.WorkRoot, "later")

	assertContains(t, resp.Output, "worktree removed:")

	// Worktree is gone
	assertFileNotExists(t, wt)

	// Later directory still exists
	assertFileExists(t, later)

	// Main repo (later) has the feature-work file (ff merged)
	assertFileExists(t, filepath.Join(later, "feature-work"))

	// History chain: [repo, mid, later] — wt spliced out
	assertHistoryChain(t, req.ConfigHome, repo, repo, mid, later)
}
```
