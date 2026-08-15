## Expected
- Exit code 0.
- Output contains "moved:" and "updated worktree:".
- repo/ no longer exists at its original location.
- wt/sub/ exists with README.md and .git dir (the main repo).
- wt/ still exists as a worktree.
- wt/.git references wt/sub (updated by moveDir to point to new main repo location).

## History
- Chain: [repo, wt(worktree), wt/sub].

## Exit Code
- 0

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	repo := filepath.Join(req.WorkRoot, "repo")
	wt := filepath.Join(req.WorkRoot, "wt")
	subDir := filepath.Join(wt, "sub")

	assertContains(t, resp.Output, "moved:")
	assertContains(t, resp.Output, "updated worktree:")

	// repo gone, sub exists as main repo
	assertFileNotExists(t, repo)
	assertFileExists(t, subDir)
	assertFileExists(t, filepath.Join(subDir, "README.md"))
	assertFileExists(t, filepath.Join(subDir, ".git"))

	// wt still exists as worktree
	assertFileExists(t, wt)
	assertFileExists(t, filepath.Join(wt, ".git"))

	// Worktree .git updated to point to sub (new main repo location)
	gitContent, err := os.ReadFile(filepath.Join(wt, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), subDir)

	assertHistoryChain(t, req.ConfigHome, repo,
		repo,
		wt,
		subDir,
	)
	assertHistoryWorktreeEntry(t, req.ConfigHome, repo, 1, repo, "wt")
}
```
