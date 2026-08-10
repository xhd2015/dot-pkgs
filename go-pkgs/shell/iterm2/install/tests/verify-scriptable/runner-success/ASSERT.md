## Expected

- `err == nil`.
- `resp.Version == "3.5.0"`.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	assertNoError(t, err)
	assertEqual(t, "Version", resp.Version, "3.5.0")
}
```
