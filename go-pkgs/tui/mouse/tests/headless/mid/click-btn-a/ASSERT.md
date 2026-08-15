---
label: needs-pty, slow
explanation: detach PTY + SGR click; expects fixture machine HIT line
---

## Expected

- `err` is nil.
- `resp.HasOrigin` true, `ViewLines == 5`.
- `resp.HasHit` true.
- `resp.HitID == "btn-a"`.
- `resp.HitLocalY == 3`.
- `resp.HitKind` is `known` (preferred) or non-empty dual kind if origin failed
  open — product should print `known` after CPR.

## Errors

- Missing ORIGIN / HIT (stub fixture).
- Wrong chip (btn-b) or wrong localY.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("click-btn-a: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if !resp.HasOrigin {
		t.Fatalf("missing ORIGIN= (snapshot:\n%s)", resp.Snapshot)
	}
	if resp.ViewLines != 5 {
		t.Fatalf("VIEW=%d, want 5", resp.ViewLines)
	}
	if !resp.HasHit {
		t.Fatalf("missing HIT= (snapshot:\n%s)", resp.Snapshot)
	}
	if resp.HitID != "btn-a" || resp.HitLocalY != 3 {
		t.Fatalf("HIT id=%q localY=%d kind=%q, want btn-a localY=3",
			resp.HitID, resp.HitLocalY, resp.HitKind)
	}
	if resp.HitKind == "" {
		t.Fatal("HIT kind empty")
	}
}
```
