## Expected
- Exit code 0.
- Output does NOT contain "worktree created:" or "worktree add".
- Destination exists with a .git file pointing to the main repo.
- Source no longer exists.

## Exit Code
- 0

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
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertNotContains(t, resp.Output, "worktree created:")
	assertNotContains(t, resp.Output, "worktree add")

	wtDst := filepath.Join(req.WorkRoot, "feature-wt-moved")
	assertFileExists(t, wtDst)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "feature-wt"))
	assertFileExists(t, filepath.Join(wtDst, ".git"))

	mainRepo := filepath.Join(req.WorkRoot, "main")
	gitContent, err := os.ReadFile(filepath.Join(wtDst, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), mainRepo)
}
```
