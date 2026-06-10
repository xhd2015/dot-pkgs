## Expected
- Exit code 0.
- Output contains "moved back".
- The original project directory exists again.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	originalPath := filepath.Join(req.WorkRoot, "projects", "myproject")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "moved back")
	assertFileExists(t, originalPath)
}
```
