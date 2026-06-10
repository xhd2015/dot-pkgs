## Expected
- Non-zero exit code.
- Output contains "not merged".
- Worktree directory still exists.

## Exit Code
- Non-zero

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\noutput: %s", resp.Output)
	}
	assertContains(t, resp.Output, "not merged")

	wtDir := filepath.Join(req.WorkRoot, "feature")
	assertFileExists(t, wtDir)
}
```
