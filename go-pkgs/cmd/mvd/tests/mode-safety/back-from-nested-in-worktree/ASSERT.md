## Expected
- Exit code 0.
- Output contains "moved back:".
- wt/sub/ no longer exists.
- repo/ exists again with README.md and .git.
- wt/ still exists as worktree.
- wt/.git now references repo (updated by moveDir during the back move).

## History
- Chain: [repo, wt(worktree)] — the wt/sub step was removed by --back.

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
	wt := filepath.Join(req.WorkRoot, "wt")
	subDir := filepath.Join(wt, "sub")

	assertContains(t, resp.Output, "moved back:")

	// sub is gone, repo restored
	assertFileNotExists(t, subDir)
	assertFileExists(t, repo)
	assertFileExists(t, filepath.Join(repo, "README.md"))
	assertFileExists(t, filepath.Join(repo, ".git"))

	// wt still exists as worktree
	assertFileExists(t, wt)
	assertFileExists(t, filepath.Join(wt, ".git"))

	// Worktree .git now references repo
	gitContent, err := os.ReadFile(filepath.Join(wt, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), repo)

	// History: wt/sub step removed
	assertHistoryChain(t, req.ConfigHome, repo,
		repo,
		wt,
	)
	assertHistoryWorktreeEntry(t, req.ConfigHome, repo, 1, repo, "wt")
}
```
