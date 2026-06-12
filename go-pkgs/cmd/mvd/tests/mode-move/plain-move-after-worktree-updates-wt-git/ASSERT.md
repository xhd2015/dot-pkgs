## Expected
- Exit code 0.
- Output contains both "moved:" and "updated worktree:".
- repo/ no longer exists.
- dst/ exists.
- wt/ still exists at its original location.
- wt/.git content references dst (the new main repo location), NOT repo.

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
	wt := filepath.Join(req.WorkRoot, "wt")

	assertContains(t, resp.Output, "moved:")
	assertContains(t, resp.Output, "updated worktree:")

	assertFileNotExists(t, repo)
	assertFileExists(t, dst)
	assertFileExists(t, wt)

	// wt/.git should reference dst (updated by moveDir)
	gitContent, err := os.ReadFile(filepath.Join(wt, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), dst)
	assertNotContains(t, string(gitContent), repo)
}
```
