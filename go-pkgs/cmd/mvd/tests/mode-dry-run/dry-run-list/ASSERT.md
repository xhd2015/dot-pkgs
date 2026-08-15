## Expected
- Exit code is 0.
- Output contains normal list output (contains the project path).
- Output does NOT contain `dry-run:` (read-only commands are unaffected).

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
	dir := filepath.Join(req.WorkRoot, "myproject")
	assertContains(t, resp.Output, filepath.Base(dir))
	assertNotContains(t, resp.Output, "dry-run:")
}
```
