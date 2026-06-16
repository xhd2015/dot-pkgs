## Expected
- Exit code 0.
- Output contains "moved back:".
- B/ no longer exists.
- A/ exists again (back from B, skipping worktree wt).
- wt/ still exists as worktree.
- wt/.git references A (updated by moveDir during B→A back).
- repo/ does NOT exist (was moved to A in step 2).

## History
- Chain after back: [repo, wt(worktree), A].
- The B step was removed by --back.

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
	A := filepath.Join(req.WorkRoot, "A")
	B := filepath.Join(req.WorkRoot, "B")
	wt := filepath.Join(req.WorkRoot, "wt")

	assertContains(t, resp.Output, "moved back:")

	// B is gone, A is back
	assertFileNotExists(t, B)
	assertFileExists(t, A)
	assertFileExists(t, filepath.Join(A, "README.md"))
	assertFileExists(t, filepath.Join(A, ".git"))

	// repo is gone (was moved to A in step 2)
	assertFileNotExists(t, repo)

	// wt still exists as worktree
	assertFileExists(t, wt)
	assertFileExists(t, filepath.Join(wt, ".git"))

	// wt/.git references A (updated by moveDir)
	gitContent, err := os.ReadFile(filepath.Join(wt, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), A)

	// History: chain is [repo, wt(wt), A] (B removed)
	assertHistoryChain(t, req.ConfigHome, repo,
		repo,
		wt,
		A,
	)
	assertHistoryWorktreeEntry(t, req.ConfigHome, repo, 1, "", "")
}
```
