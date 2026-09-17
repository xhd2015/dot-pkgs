## Expected

- No error.
- Runner called once with `["open", "https://example.com/prd"]`.
- `Result.Out` carries the runner output.

## Side Effects

- One injected runner call; no browser is opened.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertRunOnce(t, resp, []string{"open", "https://example.com/prd"})
	assertOut(t, resp, "launched")
}
```
