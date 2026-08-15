## Expected
- Exit code 0.
- Output contains "worktree removed:".
- wt/ no longer exists (git worktree remove).
- mid/ still exists (main repo not affected).
- History chain: [repo, mid] (worktree entry removed).
- No git metadata on any remaining entries.

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	repo := filepath.Join(req.WorkRoot, "repo")
	mid := filepath.Join(req.WorkRoot, "mid")
	wt := filepath.Join(req.WorkRoot, "wt")

	assertContains(t, resp.Output, "worktree removed:")

	assertFileNotExists(t, wt)
	assertFileExists(t, mid)
	assertFileExists(t, filepath.Join(mid, "README.md"))

	// History: wt removed, chain is [repo, mid]
	assertHistoryChain(t, req.ConfigHome, repo, repo, mid)
}
```
