## Expected

- `err == nil`.
- `resp.Path` is `$WorkDir/extra/mytool`.
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
	want := filepath.Join(req.ExtraDirs[0], "mytool")
	assertPathEqual(t, resp.Path, want)
	assertEqual(t, "Via", resp.Via, "extra_dir")
}
```
