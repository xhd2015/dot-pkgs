## Expected

- Replacement succeeds.
- The heading gains the default LF line ending, producing `"# Target\n"`.
- Repeated `String()` calls return the same exact bytes.

## Errors

- The returned error is nil.
- Any additional blank line or missing final LF is a failure.

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
	assertExact(t, resp.Output, "# Target\n")
	assertExact(t, resp.SecondOutput, "# Target\n")
}
```
