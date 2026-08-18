## Expected
- Non-zero exit code.
- Output indicates rebase conflict / abort (should NOT say "not merged").
- Worktree directory still exists.
- Feature branch still exists (rebase was aborted).
- Main repo does NOT have the feature change.
- Main repo still has its own change (main-work or README change).
- History unchanged.

## Exit Code
- Non-zero

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\noutput: %s", resp.Output)
	}

	// Should not be the old "not merged" error — the new rebase logic replaces it.
	assertNotContains(t, resp.Output, "not merged")

	wtDir := filepath.Join(req.WorkRoot, "feature")
	mainRepo := filepath.Join(req.WorkRoot, "main")

	// Worktree still exists
	assertFileExists(t, wtDir)

	// Main still has its README with "main change"
	data, err := os.ReadFile(filepath.Join(mainRepo, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readmeContent := string(data)
	assertContains(t, readmeContent, "main change")
	// It should NOT have the feature change
	assertNotContains(t, readmeContent, "feature change")

	// History still has the worktree entry
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```
