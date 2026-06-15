## Expected
- Exit code is 0.
- Output contains `dry-run: would move`.
- The destination `dst/src` was NOT created (no file moved inside).
- The source file `src/f.txt` still exists at its original location.

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
	assertContains(t, resp.Output, "dry-run: would move")
	src := filepath.Join(req.WorkRoot, "src")
	dstSrc := filepath.Join(req.WorkRoot, "dst", "src")
	assertFileExists(t, filepath.Join(src, "f.txt"))
	assertFileNotExists(t, dstSrc)
	assertHistoryNil(t, req.ConfigHome)
}
```
