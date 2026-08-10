## Expected

- Exactly one path: `/tmp/keep-me`.
- Empty `n`, `n/`, relative names, and non-`n` lines omitted.

## Errors

- `err` is nil.
- Including `/` or relative paths is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/tmp/keep-me"}
	assertStringsEqual(t, resp.Paths, want)
}
```
