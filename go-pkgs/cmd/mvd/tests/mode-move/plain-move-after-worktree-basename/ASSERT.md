## Expected
- Exit code 0.
- Output contains "moved:" (plain move).
- Basename "repo" resolves to the main repo (not the worktree).
- repo/ no longer exists, dst/ exists with README.md.
- wt/ still exists (was NOT moved).
- History chain: [repo, wt(worktree), dst].

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
