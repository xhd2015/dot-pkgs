---
label: needs-pty, slow
explanation: detach PTY + SGR click; btn-b exclusivity
---

## Expected

- `err` is nil.
- `resp.HasHit` true with `HitID == "btn-b"` and `HitLocalY == 4`.
- Last HIT is not `btn-a`.

## Errors

- Click mapped to btn-a (off-by-one localY / origin).
- No HIT line.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("click-btn-b-not-a: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if !resp.HasHit {
		t.Fatalf("missing HIT= (snapshot:\n%s)", resp.Snapshot)
	}
	if resp.HitID == "btn-a" {
		t.Fatalf("got HIT btn-a localY=%d; want btn-b for localY=4 click", resp.HitLocalY)
	}
	if resp.HitID != "btn-b" || resp.HitLocalY != 4 {
		t.Fatalf("HIT id=%q localY=%d kind=%q, want btn-b localY=4",
			resp.HitID, resp.HitLocalY, resp.HitKind)
	}
}
```
