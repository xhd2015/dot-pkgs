## Expected

- `err == nil`.
- `resp.Path` is `$Home/go/bin/mytool`.
- `resp.Via == "default_dir"`.

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
	want := filepath.Join(req.Home, "go", "bin", "mytool")
	assertPathEqual(t, resp.Path, want)
	assertEqual(t, "Via", resp.Via, "default_dir")
}
```
