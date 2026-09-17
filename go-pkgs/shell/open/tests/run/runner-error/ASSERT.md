## Expected

- `err != nil` and it carries the runner's message.
- No `Result` reaches the caller, so `Response.Args`/`Out` stay empty.

## Side Effects

- One runner call.

## Errors

- The injected `open: exit status 1` propagates unchanged.

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
	if !strings.Contains(err.Error(), "exit status 1") {
		t.Fatalf("err = %v, want the runner error surfaced", err)
	}
	if len(resp.Runs) != 1 {
		t.Fatalf("Runs = %#v, want exactly one attempt", resp.Runs)
	}
	if resp.Args != nil || resp.Out != "" {
		t.Fatalf("Result leaked on error: args=%#v out=%q", resp.Args, resp.Out)
	}
}
```
