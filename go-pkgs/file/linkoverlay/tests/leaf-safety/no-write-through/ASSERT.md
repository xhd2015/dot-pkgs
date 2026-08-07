## Expected

- Apply succeeds.
- `target/secret` is a regular file with content `OVERLAY`.
- `base/secret` still contains exactly `ORIGINAL` (write-through would corrupt base).

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
	leaf := targetPath(req, resp, "secret")
	st := mustLstat(t, leaf)
	if st.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("target/secret should be regular file after replace, still symlink")
	}
	assertRegularContent(t, leaf, "OVERLAY")
	assertFileContentUnchanged(t, basePath(req, "base", "secret"), "ORIGINAL")
}
```
