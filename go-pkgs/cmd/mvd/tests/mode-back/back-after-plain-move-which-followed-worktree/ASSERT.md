## Expected
- Exit code 0.
- Output contains "moved back:" (not "worktree removed").
- work/another no longer exists.
- work/base exists again with README.md.
- work/target still exists (worktree is not affected).
- target's .git file points to base (updated by moveDir).
- History chain: [base, target(wt)] (back to state after step 1, the another step removed).

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

	base := filepath.Join(req.WorkRoot, "base")
	target := filepath.Join(req.WorkRoot, "target")
	another := filepath.Join(req.WorkRoot, "another")

	assertContains(t, resp.Output, "moved back")

	assertFileNotExists(t, another)
	assertFileExists(t, base)
	assertFileExists(t, filepath.Join(base, "README.md"))

	assertFileExists(t, target)
	assertFileExists(t, filepath.Join(target, ".git"))
	assertFileExists(t, filepath.Join(target, "README.md"))

	assertHistoryChain(t, req.ConfigHome, base, base, target)
	assertHistoryWorktreeEntry(t, req.ConfigHome, base, 1, base, "target")
}
```
