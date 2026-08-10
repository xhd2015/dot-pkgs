## Expected

- `err == nil` preferred (unknown latest → skip update; no hard error required).
- `resp.Action == "noop"`.
- `len(resp.ShellCalls) == 0` (no install, no update).
- `resp.FetchLatestCalls >= 1`.

## Errors

- Optional soft note allowed; must **not** force install when bin is present.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	// Prefer nil error; if product returns error it must still not install.
	// Requirement: Action=noop (skip update), no install.
	if err != nil {
		// Allow soft error only if still noop with no mutators.
		if resp.Action != "noop" && resp.Action != "" {
			t.Fatalf("on latest-fail err=%v Action=%q, want noop", err, resp.Action)
		}
	} else {
		assertEqual(t, "Action", resp.Action, "noop")
	}
	if len(resp.ShellCalls) != 0 {
		t.Fatalf("ShellCalls = %#v, want empty (no install/update)", resp.ShellCalls)
	}
	if resp.FetchLatestCalls < 1 {
		t.Fatalf("FetchLatestCalls = %d, want >= 1", resp.FetchLatestCalls)
	}
}
```
