## Expected
- `mvd src dst` moves the src directory into dst (renamed to dst).
- The output contains "moved:".
- The file that was in src is now in dst.
- The history records one project with the move chain.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	src := filepath.Join(req.WorkRoot, "src")
	dst := filepath.Join(req.WorkRoot, "dst")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "moved:")
	assertFileExists(t, filepath.Join(dst, "f.txt"))
	assertFileNotExists(t, src)
	assertHistoryChain(t, req.ConfigHome, src, src, dst)
}
```
