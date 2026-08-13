## Expected

- `err != nil` and the message contains the injected `OpenITerm` error (or a wrap of it).
- `OpenITerm` is called once with `req.Dir`.
- `OpenTerminal` is **not** called (no fallthrough).

## Side Effects

- iTerm opener runs once; Terminal opener does not run.

## Errors

- The `OpenITerm` error is returned (or wrapped). Not a Terminal path.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertError(t, err)
	if !strings.Contains(err.Error(), req.OpenITermErr) {
		t.Fatalf("error %q does not contain injected OpenITerm message %q", err.Error(), req.OpenITermErr)
	}
	assertOpenITermOnce(t, resp, req.Dir)
	assertNoOpenTerminal(t, resp)
}
```
