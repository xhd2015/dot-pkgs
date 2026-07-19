---
label: needs-pty, slow
explanation: top-anchored PTY geometry + click
---

## Expected

- `err` is nil.
- `resp.HasOrigin` true; for top anchor `OriginY` is small (`<= 2`).
- `resp.HitID == "btn-a"` and `HitLocalY == 3`.

## Errors

- Mid/bottom origin when anchor=top.
- Wrong HIT id.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("top/click-btn-a: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if !resp.HasOrigin {
		t.Fatalf("missing ORIGIN= (snapshot:\n%s)", resp.Snapshot)
	}
	if resp.OriginY > 2 {
		t.Fatalf("top ORIGIN=%d, want <= 2", resp.OriginY)
	}
	if !resp.HasHit || resp.HitID != "btn-a" || resp.HitLocalY != 3 {
		t.Fatalf("HIT id=%q localY=%d kind=%q, want btn-a localY=3",
			resp.HitID, resp.HitLocalY, resp.HitKind)
	}
}
```
