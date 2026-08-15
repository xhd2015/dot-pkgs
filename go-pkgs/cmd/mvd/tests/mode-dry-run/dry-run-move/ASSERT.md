## Expected
- Exit code is 0 (no errors).
- Output contains `dry-run: would move`.
- The source directory `src` and file `src/f.txt` still exist (not moved).
- The destination directory `dst` was NOT created.
- No history was written.

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "dry-run: would move")
	src := filepath.Join(req.WorkRoot, "src")
	dst := filepath.Join(req.WorkRoot, "dst")
	assertFileExists(t, filepath.Join(src, "f.txt"))
	assertFileNotExists(t, dst)
	assertHistoryNil(t, req.ConfigHome)
}
```
