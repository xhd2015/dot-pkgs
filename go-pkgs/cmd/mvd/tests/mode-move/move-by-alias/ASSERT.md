## Expected
- mvd resolves the alias "kk" to the tracked project "kool".
- The project ends up at final/kool after the second move.
- History records the original root plus the two move destinations.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	src := filepath.Join(req.WorkRoot, "projects", "kool")
	scratch := filepath.Join(req.WorkRoot, "scratch", "kool")
	final := filepath.Join(req.WorkRoot, "final", "kool")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertFileExists(t, final)
	assertFileNotExists(t, scratch)
	assertHistoryChain(t, req.ConfigHome, src, src, scratch, final)
}
```
