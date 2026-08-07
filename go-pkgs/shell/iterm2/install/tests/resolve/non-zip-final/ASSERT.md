## Expected

- `err != nil` (final URL is not a zip).

## Errors

- Non-nil error from `ResolveLatestStableURL`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	_ = resp
	assertError(t, err)
}
```
