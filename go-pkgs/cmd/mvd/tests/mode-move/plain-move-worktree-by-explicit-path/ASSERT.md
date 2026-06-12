## Expected
- Exit code 0.
- Output contains "moved" (plain move of the worktree).
- repo/ still exists at its original location (main repo was NOT moved).
- wt/ no longer exists (it was moved to dst).
- dst/ exists with .git file still pointing to repo.

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

	assertContains(t, resp.Output, "moved")
	assertNotContains(t, resp.Output, "worktree created:")

	// Main repo should NOT have been moved
	assertFileExists(t, repo)
	assertFileExists(t, filepath.Join(repo, "README.md"))

	// Worktree was moved to dst
	assertFileNotExists(t, wt)
	assertFileExists(t, dst)
	assertFileExists(t, filepath.Join(dst, ".git"))

	// History: [repo, wt, dst] — wt was moved (not repo)
	assertHistoryChain(t, req.ConfigHome, repo, repo, wt, dst)
}
```
