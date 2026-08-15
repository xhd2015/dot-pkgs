## Expected
- Exit code 0.
- Output contains "moved:" (plain move), NOT "worktree created:".
- repo/ no longer exists at its original location.
- dst/ exists with README.md (the main repo was moved there).
- wt/ still exists at its original worktree location (it was NOT moved).
- History chain: [repo, wt(worktree), dst].

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
	dst := filepath.Join(req.WorkRoot, "dst")
	wt := filepath.Join(req.WorkRoot, "wt")

	assertContains(t, resp.Output, "moved:")
	assertNotContains(t, resp.Output, "worktree created:")

	assertFileNotExists(t, repo)
	assertFileExists(t, dst)
	assertFileExists(t, filepath.Join(dst, "README.md"))

	assertFileExists(t, wt)
	assertFileExists(t, filepath.Join(wt, ".git"))
	assertFileExists(t, filepath.Join(wt, "README.md"))

	assertHistoryChain(t, req.ConfigHome, repo, repo, wt, dst)
	assertHistoryWorktreeEntry(t, req.ConfigHome, repo, 1, repo, "wt")
}
```
