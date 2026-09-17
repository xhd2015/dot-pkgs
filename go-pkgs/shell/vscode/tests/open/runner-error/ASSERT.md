## Expected

- `err != nil` carrying the runner's message.
- No `Result`, so `Response.CodePath`/`Args` stay empty.
- The runner was attempted exactly once.

## Side Effects

- One resolve call and one runner call.

## Errors

- The runner error propagates unchanged.

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
	if resp.Args != nil || resp.CodePath != "" {
		t.Fatalf("Result leaked on error: args=%#v codePath=%q", resp.Args, resp.CodePath)
	}
}
```
