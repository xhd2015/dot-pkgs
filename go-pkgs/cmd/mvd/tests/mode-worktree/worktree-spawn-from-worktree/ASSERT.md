## Expected
- Exit code 0.
- Output contains "worktree created:" and "[branch: wt2]".
- work/wt2 exists with a `.git` regular file (linked worktree layout).
- History chain is keyed on wt1 (the worktree used as SRC).
- Worktree metadata records MainRepo as wt1.

## Exit Code
- 0

```go
import (
	"os"
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
	assertContains(t, resp.Output, "[branch: wt2]")

	wt1 := filepath.Join(req.WorkRoot, "wt1")
	wt2 := filepath.Join(req.WorkRoot, "wt2")
	assertFileExists(t, wt2)

	gitInfo, err := os.Stat(filepath.Join(wt2, ".git"))
	assertErrIsNil(t, err)
	if gitInfo.IsDir() {
		t.Fatalf("expected wt2/.git to be a regular file (worktree), got directory")
	}

	assertHistoryChain(t, req.ConfigHome, wt1, wt1, wt2)
	assertHistoryWorktreeEntry(t, req.ConfigHome, wt1, 1, wt1, "wt2")
}
```