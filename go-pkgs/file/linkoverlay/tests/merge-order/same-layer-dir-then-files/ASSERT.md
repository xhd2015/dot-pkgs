## Expected

- Apply succeeds.
- `target/marker` content is `B` (Files after Dir).
- Base `base/marker` remains `A` (no write-through).
- `target/marker` is a regular file (not a symlink into base), because the leaf
  was replaced by the Files write.

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
	_ = d
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	path := targetPath(req, resp, "marker")
	st := mustLstat(t, path)
	if st.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("target/marker should be a regular file after Files overlay, still symlink")
	}
	assertRegularContent(t, path, "B")
	assertFileContentUnchanged(t, basePath(req, "base", "marker"), "A")
}
```
