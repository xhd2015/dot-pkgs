---
label: needs-pty, slow
explanation: bottom-anchored PTY geometry + click
---

## Expected

- `err` is nil.
- `resp.HasOrigin` true; for bottom anchor `OriginY` is high (`>= 10` on
  typical 24-row PTY with VIEW=5).
- `resp.HitID == "btn-a"` and `HitLocalY == 3`.

## Errors

- Top origin when anchor=bottom.
- Wrong HIT id.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("bottom/click-btn-a: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if !resp.HasOrigin {
		t.Fatalf("missing ORIGIN= (snapshot:\n%s)", resp.Snapshot)
	}
	if resp.OriginY < 10 {
		t.Fatalf("bottom ORIGIN=%d, want >= 10", resp.OriginY)
	}
	if !resp.HasHit || resp.HitID != "btn-a" || resp.HitLocalY != 3 {
		t.Fatalf("HIT id=%q localY=%d kind=%q, want btn-a localY=3",
			resp.HitID, resp.HitLocalY, resp.HitKind)
	}
}
```
