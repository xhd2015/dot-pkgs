## Expected
- Exit code 0.
- Output contains "moved:".
- repo/ no longer exists (moved to dst).
- dst/ exists with README.md.
- wt1/ and wt2/ still exist at their original worktree locations.
- wt1/.git and wt2/.git reference dst (updated via moveDir).
- History chain: [repo, wt1(worktree), wt2(worktree), dst].

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
	dst := filepath.Join(req.WorkRoot, "dst")
	wt1 := filepath.Join(req.WorkRoot, "wt1")
	wt2 := filepath.Join(req.WorkRoot, "wt2")

	assertContains(t, resp.Output, "moved:")
	assertNotContains(t, resp.Output, "worktree created:")

	assertFileNotExists(t, repo)
	assertFileExists(t, dst)
	assertFileExists(t, filepath.Join(dst, "README.md"))

	assertFileExists(t, wt1)
	assertFileExists(t, filepath.Join(wt1, ".git"))

	assertFileExists(t, wt2)
	assertFileExists(t, filepath.Join(wt2, ".git"))

	// Both worktree .git files should reference dst
	gitContent1, err := os.ReadFile(filepath.Join(wt1, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent1), dst)

	gitContent2, err := os.ReadFile(filepath.Join(wt2, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent2), dst)

	assertHistoryChain(t, req.ConfigHome, repo, repo, wt1, wt2, dst)
	assertHistoryWorktreeEntry(t, req.ConfigHome, repo, 1, repo, "")
	assertHistoryWorktreeEntry(t, req.ConfigHome, repo, 2, repo, "")
}
```
