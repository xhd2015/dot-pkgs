## Expected

- `err == nil`.
- `resp.Path` equals the injected LookPath hit.
- `resp.Via == "path"`.
- `resp.LookPathCalls` contains `"mytool"`.

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
	assertPathEqual(t, resp.Path, req.LookPathHit)
	assertEqual(t, "Via", resp.Via, "path")
	if len(resp.LookPathCalls) == 0 {
		t.Fatal("expected LookPath to be called for bare name")
	}
	assertEqual(t, "LookPathCalls[0]", resp.LookPathCalls[0], req.Name)
}
```
