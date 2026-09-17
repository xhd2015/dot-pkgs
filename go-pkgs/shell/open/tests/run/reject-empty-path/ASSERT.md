## Expected

- `err != nil`.
- The injected runner is not called.

## Side Effects

- No process launch.

## Errors

- `open: target is empty` (whitespace-only counts as empty).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertError(t, err)
	assertNoRun(t, resp)
}
```
