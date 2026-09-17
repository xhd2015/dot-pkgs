## Expected

- No error.
- Resolve called once with `Home`.
- Runner called once with `[<CodePath>, "-r", <Dir>]`.
- `Result.CodePath` and `Result.Via` mirror what the resolver reported.

## Side Effects

- One resolve call and one runner call; no window is opened.

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
	assertResolvedHome(t, resp, req.Home)
	assertDefaultRunner(t, resp, req, []string{req.CodePath, "-r", req.Dir})
	if resp.Via != "candidate" {
		t.Fatalf("Result.Via = %q, want candidate", resp.Via)
	}
}
```
