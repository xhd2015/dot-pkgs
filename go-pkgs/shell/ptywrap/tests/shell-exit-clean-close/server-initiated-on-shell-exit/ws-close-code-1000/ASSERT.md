## Expected

1. `CloseCode` is **1000** (`websocket.CloseNormalClosure`).
2. Session remains listable after exit (`SessionListed` true) when the server
   keeps exited metadata (existing lifecycle).
3. If status is reported, `SessionStatus` is `exited` (best-effort).

## Errors

- `CloseCode` is **1006** (abnormal / unexpected EOF) — bare `conn.Close()` bug.
- Any non-1000 close code from server-initiated end after shell exit.
- Harness timeout waiting for server close (child never exited / attach too late).

## Side Effects

- Test cleanup deletes the session via REST.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/gorilla/websocket"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.CloseCode != websocket.CloseNormalClosure {
		t.Fatalf("CloseCode=%d, want %d (Normal Closure); 1006 means bare conn.Close / unexpected EOF",
			resp.CloseCode, websocket.CloseNormalClosure)
	}
	if !resp.SessionListed {
		t.Fatalf("session %s not listed after shell exit (expected exited metadata retained)",
			resp.SessionID)
	}
	if resp.SessionStatus != "" && resp.SessionStatus != "exited" {
		t.Fatalf("SessionStatus=%q, want exited (or empty if not reported)", resp.SessionStatus)
	}
}
```
