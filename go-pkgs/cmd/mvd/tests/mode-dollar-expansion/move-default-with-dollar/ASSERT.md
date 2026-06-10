## Expected
- Exit code 0.
- Output contains "moved:".
- The file is in dst, original directory no longer exists.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	originalPath := filepath.Join(req.WorkRoot, "projects", "myproject")
	dst := filepath.Join(req.WorkRoot, "dst")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "moved:")
	assertFileExists(t, filepath.Join(dst, "f.txt"))
	assertFileNotExists(t, originalPath)
}
```
