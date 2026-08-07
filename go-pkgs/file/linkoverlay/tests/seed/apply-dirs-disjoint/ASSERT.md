## Expected

- `ApplyDirs` succeeds (`err == nil`).
- `target/a.txt` is an absolute symlink to `base-a/a.txt`; content reads `from-a`.
- `target/b.txt` is an absolute symlink to `base-b/b.txt`; content reads `from-b`.
- `target/.config` is an absolute symlink to `base-a/.config`; nested `tool` reads `cfg-a`.

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
		t.Fatalf("ApplyDirs: %v", err)
	}
	assertSymlinkTo(t, targetPath(req, resp, "a.txt"), basePath(req, "base-a", "a.txt"))
	assertSymlinkTo(t, targetPath(req, resp, "b.txt"), basePath(req, "base-b", "b.txt"))
	assertSymlinkTo(t, targetPath(req, resp, ".config"), basePath(req, "base-a", ".config"))

	assertRegularContent(t, targetPath(req, resp, "a.txt"), "from-a")
	assertRegularContent(t, targetPath(req, resp, "b.txt"), "from-b")
	assertRegularContent(t, targetPath(req, resp, ".config/tool"), "cfg-a")
}
```
