## Expected
- Exit code 0.
- Destination (pricing-moved) exists with a .git file pointing to the main repo (not the intermediate worktree).
- Source (pricing) no longer exists.

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

	wt2Dst := filepath.Join(req.WorkRoot, "pricing-moved")
	assertFileExists(t, wt2Dst)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "pricing"))
	assertFileExists(t, filepath.Join(wt2Dst, ".git"))

	mainRepo := filepath.Join(req.WorkRoot, "main")
	gitContent, err := os.ReadFile(filepath.Join(wt2Dst, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), mainRepo)
}
```
