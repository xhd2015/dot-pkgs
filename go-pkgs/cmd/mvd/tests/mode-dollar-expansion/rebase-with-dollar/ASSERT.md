## Expected
- Exit code 0.
- Output contains "rebased:".
- History chain reflects the rebase with the new base as key.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "rebased:")
	newBase := filepath.Join(req.WorkRoot, "rebased")
	originalPath := filepath.Join(req.WorkRoot, "projects", "myproject")
	d1Path := filepath.Join(req.WorkRoot, "d1", "myproject")
	assertHistoryChain(t, req.ConfigHome, newBase, newBase, originalPath, d1Path)
}
```
