## Expected
- Exit code 0 (operation aborted cleanly after decline).
- Output lists concrete planned git commands before the confirmation prompt.
- Output contains `git -C`, `rebase`, `merge --ff-only`, `worktree remove`, `branch -D feature`, and `Proceed? [Y/n]`.
- Worktree directory still exists.
- Main repo does NOT have the feature change; main still has its own change.
- History unchanged (worktree entry still present).

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	wtDir := filepath.Join(req.WorkRoot, "feature")
	mainRepo := filepath.Join(req.WorkRoot, "main")

	assertContains(t, resp.Output, "git -C")
	assertContains(t, resp.Output, "rebase")
	assertContains(t, resp.Output, "merge --ff-only")
	assertContains(t, resp.Output, "worktree remove")
	assertContains(t, resp.Output, wtDir)
	assertContains(t, resp.Output, "branch -D feature")
	assertContains(t, resp.Output, "Proceed? [Y/n]")

	assertFileExists(t, wtDir)
	assertFileExists(t, filepath.Join(wtDir, ".git"))
	assertFileNotExists(t, filepath.Join(mainRepo, "feature-work"))
	assertFileExists(t, filepath.Join(mainRepo, "main-work"))
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```