## Expected
- Exit code 0.
- Output contains "moved:".
- repo/, A/, B/ no longer exist.
- dst/ exists with README.md (main repo moved to dst).
- wt/ still exists (worktree NOT moved).
- wt/.git references dst (updated via moveDir).
- History chain: [repo, A, B, wt(worktree), dst].

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
	a := filepath.Join(req.WorkRoot, "A")
	b := filepath.Join(req.WorkRoot, "B")
	dst := filepath.Join(req.WorkRoot, "dst")
	wt := filepath.Join(req.WorkRoot, "wt")

	assertContains(t, resp.Output, "moved:")
	assertNotContains(t, resp.Output, "worktree created:")

	assertFileNotExists(t, repo)
	assertFileNotExists(t, a)
	assertFileNotExists(t, b)
	assertFileExists(t, dst)
	assertFileExists(t, filepath.Join(dst, "README.md"))

	assertFileExists(t, wt)
	assertFileExists(t, filepath.Join(wt, ".git"))

	gitContent, err := os.ReadFile(filepath.Join(wt, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), dst)

	assertHistoryChain(t, req.ConfigHome, repo, repo, a, b, wt, dst)
	assertHistoryWorktreeEntry(t, req.ConfigHome, repo, 3, b, "")
}
```
