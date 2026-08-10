## Expected

- Apply succeeds.
- `target/pack` is regular file content `content-C` (Files last).
- `target/extra` readable as `from-B` (from later Dir layer).
- Bases `base-a/pack` and `base-b/pack` unchanged on disk.

## Errors

- No error.

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	pack := targetPath(req, resp, "pack")
	st := mustLstat(t, pack)
	if st.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("target/pack should be regular file after Files overlay")
	}
	assertRegularContent(t, pack, "content-C")
	assertRegularContent(t, targetPath(req, resp, "extra"), "from-B")
	assertFileContentUnchanged(t, basePath(req, "base-a", "pack"), "content-A")
	assertFileContentUnchanged(t, basePath(req, "base-b", "pack"), "content-B")
}
```
