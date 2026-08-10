## Expected

- `err == nil`.
- `resp.Path` is the executable under the **second** ExtraDir.
- `resp.Via == "extra_dir"`.

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	want := filepath.Join(req.ExtraDirs[1], "mytool")
	assertPathEqual(t, resp.Path, want)
	assertEqual(t, "Via", resp.Via, "extra_dir")
}
```
