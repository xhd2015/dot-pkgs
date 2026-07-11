## Expected

1. `err` is nil and `resp` is non-nil.
2. `ProcessAlive` is true after ≥3 snapshot attaches.
3. `SnapshotCount` ≥ `RepeatCount` (default 3) with non-empty frames.
4. Final `WSOutput` contains sticky marker (`STICKY_FOOTER`).
5. `SessionListed` is true (session not deleted by snapshot disconnect).

## Errors

- Child dead → snapshot incorrectly claimed writer / `stopChild` on close.
- Empty or fewer snapshots → frame delivery regression.
- Sticky missing → live-screen export broken on multi-poll path.

## Side Effects

- Cleanup kills leftover fixture PID and deletes session.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	wantN := req.RepeatCount
	if wantN <= 0 {
		wantN = 3
	}
	if !resp.ProcessAlive {
		t.Fatalf("child dead after %d snapshot attaches (session=%s) — snapshot must not stopChild",
			wantN, resp.SessionID)
	}
	if resp.SnapshotCount < wantN {
		t.Fatalf("SnapshotCount=%d, want >= %d usable non-empty snapshots",
			resp.SnapshotCount, wantN)
	}
	sticky := req.StickyMarker
	if sticky == "" {
		sticky = "STICKY_FOOTER"
	}
	if !strings.Contains(resp.WSOutput, sticky) {
		t.Fatalf("WSOutput missing sticky %q after multi-snapshot (got %q)",
			sticky, truncate(resp.WSOutput, 240))
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
