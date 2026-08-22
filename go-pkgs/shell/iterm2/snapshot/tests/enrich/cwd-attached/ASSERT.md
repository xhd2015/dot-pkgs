## Expected

- Capture succeeds; session is idle.
- `session.Cwd` is `/Users/me/proj` (from ListCwds inject).

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	snap := mustSnap(t, resp, err)
	s := snap.Windows[0].Tabs[0].Sessions[0]
	idle, ok := boolVal(s.Idle)
	if !ok || !idle {
		t.Fatalf("want idle session for cwd case, got %#v", s.Idle)
	}
	if strVal(s.Cwd) != "/Users/me/proj" {
		t.Fatalf("Cwd=%q want /Users/me/proj", strVal(s.Cwd))
	}
}
```
