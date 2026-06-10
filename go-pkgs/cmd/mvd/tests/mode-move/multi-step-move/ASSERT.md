## Expected
- After two consecutive moves (src -> d1, then d1/src -> d2), the directory ends up at d2/src.
- The history records the full chain of locations (original root plus all intermediate and final locations).

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
	assertFileExists(t, d2)
	assertFileNotExists(t, filepath.Join(d1, "src"))
	assertHistoryChain(t, req.ConfigHome, src, src, filepath.Join(d1, "src"), filepath.Join(d2, "src"))
}
```
