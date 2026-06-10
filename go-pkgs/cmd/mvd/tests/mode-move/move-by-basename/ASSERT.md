## Expected
- mvd resolves "kool" to the tracked project by its basename.
- The project is moved to archive/kool.
- History records the move from the original location.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	src := filepath.Join(req.WorkRoot, "projects", "kool")
	dst := filepath.Join(req.WorkRoot, "archive", "kool")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertFileExists(t, dst)
	assertFileNotExists(t, src)
	assertHistoryChain(t, req.ConfigHome, src, src, dst)
}
```
