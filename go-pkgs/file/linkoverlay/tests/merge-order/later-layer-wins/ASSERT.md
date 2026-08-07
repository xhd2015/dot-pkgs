## Expected

- Apply succeeds.
- Reading `target/marker` yields `late` (later Dir layer).
- On-disk `base0/marker` remains `early` (no mutation of earlier base).
- Prefer `target/marker` as abs symlink to `base1/marker` when still a seed link.

## Errors

- No error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	assertRegularContent(t, targetPath(req, resp, "marker"), "late")
	assertFileContentUnchanged(t, basePath(req, "base0", "marker"), "early")
	assertFileContentUnchanged(t, basePath(req, "base1", "marker"), "late")
}
```
