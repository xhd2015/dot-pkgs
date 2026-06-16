## Expected
- Non-zero exit code.
- Output mentions `--confirm-from-stdin`.
- Worktree still exists; feature commit not merged into main.

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

	assertContains(t, resp.Output, "--confirm-from-stdin")

	wtDir := filepath.Join(req.WorkRoot, "feature")
	mainRepo := filepath.Join(req.WorkRoot, "main")

	assertFileExists(t, wtDir)
	assertFileNotExists(t, filepath.Join(mainRepo, "feature-work"))
}
```