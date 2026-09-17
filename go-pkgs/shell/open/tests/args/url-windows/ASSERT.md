## Expected

- No error.
- `Args` = `["cmd", "/c", "start", "", "https://example.com/prd"]`.
- The empty argument sits before the URL: `start` would otherwise take the
  first quoted token as the window title.

## Side Effects

- None: `URLArgs` is pure.

## Errors

- None for a non-empty URL on windows.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertArgs(t, resp.Args, []string{"cmd", "/c", "start", "", "https://example.com/prd"})
}
```
