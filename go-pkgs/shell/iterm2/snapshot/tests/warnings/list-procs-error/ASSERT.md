## Expected

- Capture returns a snapshot (hierarchy present) with `err == nil`.
- Session is unknown (no usable procs) or at least not hard-failed.
- Warnings non-empty; at least one contains `warning` and either `ps` or the
  inject error text / tty.

## Errors

- Hard `err` must be nil (soft path only).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	snap := mustSnap(t, resp, err)
	if len(snap.Windows) != 1 {
		t.Fatalf("expected hierarchy preserved, windows=%d", len(snap.Windows))
	}
	if len(resp.Warnings) == 0 {
		t.Fatal("expected soft warnings when ListProcs fails")
	}
	joined := strings.ToLower(strings.Join(resp.Warnings, "\n"))
	if !strings.Contains(joined, "warning") {
		t.Fatalf("warnings should be soft-path style: %v", resp.Warnings)
	}
	if !strings.Contains(joined, "ps") && !strings.Contains(joined, "failed") && !strings.Contains(joined, "ttys005") {
		t.Fatalf("warnings missing probe failure signal: %v", resp.Warnings)
	}
}
```
