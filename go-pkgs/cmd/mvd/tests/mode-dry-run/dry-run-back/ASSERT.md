## Expected
- Exit code is 0.
- Output contains `dry-run: would move back`.
- The file `src/f.txt` is still at the moved location `dst/src/f.txt`.
- The history chain still has 2 entries (not reduced).

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
	assertContains(t, resp.Output, "dry-run: would move back")
	src := filepath.Join(req.WorkRoot, "src")
	dstSrc := filepath.Join(req.WorkRoot, "dst", "src")
	assertFileExists(t, filepath.Join(dstSrc, "f.txt"))
	assertFileNotExists(t, src)
	assertHistoryChain(t, req.ConfigHome, src, src, dstSrc)
}
```
