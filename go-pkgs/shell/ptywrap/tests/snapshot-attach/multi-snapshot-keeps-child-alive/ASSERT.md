## Expected

1. `ProcessAlive` is true after ≥3 snapshot attaches.
2. `SnapshotCount` is ≥ 3 (all attaches returned non-empty bytes).
3. `WSOutput` contains `SNAP-MARKER` (usable scrollback frame).
4. `SessionListed` is true (session not deleted by snapshot disconnect).

## Errors

- Child dead after multi-snapshot (attach incorrectly claimed writer / stopChild).
- Empty snapshots (frame not delivered).

## Side Effects

- Test cleanup kills leftover sleep child and deletes session.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if !resp.ProcessAlive {
		t.Fatalf("child dead after %d snapshot attaches (session=%s) — snapshot mode must not stopChild on disconnect",
			req.RepeatCount, resp.SessionID)
	}
	wantN := req.RepeatCount
	if wantN <= 0 {
		wantN = 3
	}
	if resp.SnapshotCount < wantN {
		t.Fatalf("SnapshotCount=%d, want >= %d usable non-empty snapshots",
			resp.SnapshotCount, wantN)
	}
	if !strings.Contains(resp.WSOutput, "SNAP-MARKER") {
		t.Fatalf("WSOutput missing SNAP-MARKER (got %q)", truncate(resp.WSOutput, 200))
	}
	if !resp.SessionListed {
		t.Fatalf("session %s not listed after snapshot attaches", resp.SessionID)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
