## Expected
- Moving by root path resolves to the current location and moves it to the new destination.
- The file ends up at d2/src, not at d1/src.
- History captures the full chain including the original root and all intermediate locations.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	src := filepath.Join(req.WorkRoot, "src")
	d1 := filepath.Join(req.WorkRoot, "d1")
	d2 := filepath.Join(req.WorkRoot, "d2")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertFileExists(t, filepath.Join(d2, "src"))
	assertFileNotExists(t, filepath.Join(d1, "src"))
	assertHistoryChain(t, req.ConfigHome, src, src, filepath.Join(d1, "src"), filepath.Join(d2, "src"))
}
```
