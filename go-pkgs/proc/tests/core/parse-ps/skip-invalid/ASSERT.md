## Expected

- Exactly one row: `{PID:7, PPID:1, Cmd:"/bin/true"}`.
- No panic; invalid lines omitted.

## Errors

- `err` is nil.
- Extra rows from garbage or empty result is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	want := []FixtureProc{
		{PID: 7, PPID: 1, Cmd: "/bin/true"},
	}
	assertProcsEqual(t, resp.Procs, want)
}
```
