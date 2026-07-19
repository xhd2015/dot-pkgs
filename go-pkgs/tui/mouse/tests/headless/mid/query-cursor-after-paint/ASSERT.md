---
label: needs-pty, slow
explanation: detach PTY session + CSI 6n origin; fixture product not complete yet
---

## Expected

- `err` is nil (session started and query-cursor returned).
- Either:
  - `resp.HasOrigin` is true with `ViewLines == 5`, and mid origin is
    reasonably mid-pane (`OriginY >= 4`), **or**
  - `resp.HasCursor` is true and cursor row is near the last painted UI line
    when origin is known: `abs(CursorRow - (OriginY+ViewLines-1)) <= 1`.
- Snapshot contains `fixture-inline` or `btn-a` or `ORIGIN=`.

## Errors

- No ORIGIN and no usable cursor after paint.
- Origin claims VIEW other than 5 when ORIGIN is present.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("query-cursor-after-paint: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	snapOK := strings.Contains(resp.Snapshot, "ORIGIN=") ||
		strings.Contains(resp.Snapshot, "fixture-inline") ||
		strings.Contains(resp.Snapshot, "btn-a")
	if !snapOK {
		t.Fatalf("snapshot missing UI/ORIGIN markers:\n%s", resp.Snapshot)
	}

	if resp.HasOrigin {
		if resp.ViewLines != 5 {
			t.Fatalf("VIEW=%d, want 5", resp.ViewLines)
		}
		// mid pad ~8; allow some slack once CPR works
		if resp.OriginY < 4 {
			t.Fatalf("mid ORIGIN=%d, want >= 4 (mid-pane)", resp.OriginY)
		}
		if resp.HasCursor {
			want := resp.OriginY + resp.ViewLines - 1
			delta := resp.CursorRow - want
			if delta < 0 {
				delta = -delta
			}
			if delta > 1 {
				t.Fatalf("cursor row=%d, want near last UI line %d (±1)", resp.CursorRow, want)
			}
		}
		return
	}

	// ORIGIN optional if host query-cursor already shows a plausible mid-pane row.
	if !resp.HasCursor {
		t.Fatal("neither ORIGIN= nor query-cursor result present")
	}
	if resp.CursorRow < 4 {
		t.Fatalf("mid query-cursor row=%d, want mid-pane-ish row >= 4 (and/or ORIGIN= line)", resp.CursorRow)
	}
}
```
