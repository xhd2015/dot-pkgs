## Expected

- `err != nil` and the message contains the injected `OpenTerminal` error (or a wrap of it).
- `OpenTerminal` is called once with `req.Dir`.
- `OpenITerm` is not called.

## Side Effects

- Terminal opener runs once; iTerm opener does not run.

## Errors

- The `OpenTerminal` error is returned (or wrapped).

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
	if !strings.Contains(err.Error(), req.OpenTerminalErr) {
		t.Fatalf("error %q does not contain injected OpenTerminal message %q", err.Error(), req.OpenTerminalErr)
	}
	assertOpenTerminalOnce(t, resp, req.Dir)
	assertNoOpenITerm(t, resp)
}
```
