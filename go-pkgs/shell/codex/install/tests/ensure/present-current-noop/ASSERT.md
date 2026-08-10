## Expected

- `err == nil`.
- `resp.Action == "noop"`.
- `len(resp.ShellCalls) == 0` (no install/update).
- `resp.FetchLatestCalls >= 1`.
- `resp.ResultNeedsUpdate == false`.

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
	_ = req
	assertNoError(t, err)
	assertEqual(t, "Action", resp.Action, "noop")
	if len(resp.ShellCalls) != 0 {
		t.Fatalf("ShellCalls = %#v, want empty (noop)", resp.ShellCalls)
	}
	if resp.FetchLatestCalls < 1 {
		t.Fatalf("FetchLatestCalls = %d, want >= 1", resp.FetchLatestCalls)
	}
	assertEqual(t, "ResultNeedsUpdate", resp.ResultNeedsUpdate, false)
}
```
