## Expected
- Exit code 0.
- Output contains "moved:".
- repo/ and mid/ no longer exist.
- dst/ exists with README.md (main repo moved to dst).
- wt/ still exists (worktree was NOT moved).
- wt/.git references dst (worktree link updated via moveDir).
- History chain: [repo, mid, wt(worktree), dst].

## Exit Code
- 0

```go
import (
	"os"
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	repo := filepath.Join(req.WorkRoot, "repo")
	mid := filepath.Join(req.WorkRoot, "mid")
	dst := filepath.Join(req.WorkRoot, "dst")
	wt := filepath.Join(req.WorkRoot, "wt")

	assertContains(t, resp.Output, "moved:")
	assertNotContains(t, resp.Output, "worktree created:")

	assertFileNotExists(t, repo)
	assertFileNotExists(t, mid)
	assertFileExists(t, dst)
	assertFileExists(t, filepath.Join(dst, "README.md"))

	// Worktree should still exist
	assertFileExists(t, wt)
	assertFileExists(t, filepath.Join(wt, ".git"))

	// Worktree .git should reference the new main repo location (dst)
	gitContent, err := os.ReadFile(filepath.Join(wt, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), dst)

	// History chain
	assertHistoryChain(t, req.ConfigHome, repo, repo, mid, wt, dst)
	assertHistoryWorktreeEntry(t, req.ConfigHome, repo, 2, mid, "wt")
}
```
