## Expected

- Final `err == nil` and `LastStopErr == nil`.
- State Hosts is empty (len 0).
- `State.ConnectorPID == 0` (connector down; no logical run left alive).
- Config, when present, has **no** hostname rules; only 404 catch-all is OK
  (Ingress length 0 or 1 with empty hostname + `http_status:404`).

## Side Effects

- Managed state.json / config.yml rewritten; connector sentinel cleared.

## Errors

- None.

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Attach/Stop error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.LastStopErr != nil {
		t.Fatalf("Stop(A) error: %v", resp.LastStopErr)
	}
	if resp.State == nil {
		t.Fatal("nil State after last-host Stop")
	}
	if len(resp.State.Hosts) != 0 {
		t.Fatalf("Hosts len = %d, want 0; Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	if resp.State.ConnectorPID != 0 {
		t.Fatalf("ConnectorPID = %d, want 0 after last host Stop", resp.State.ConnectorPID)
	}

	if resp.Config != nil {
		svcs := hostRuleServices(resp.Config)
		if len(svcs) != 0 {
			t.Fatalf("host rules remain after last Stop: %v", svcs)
		}
		// 404-only or empty ingress both acceptable.
		if len(resp.Config.Ingress) > 1 {
			t.Fatalf("Ingress len = %d, want 0 or 1 (404-only)", len(resp.Config.Ingress))
		}
		if len(resp.Config.Ingress) == 1 {
			r := resp.Config.Ingress[0]
			if r.Hostname != "" || r.Service != "http_status:404" {
				t.Fatalf("sole ingress = %+v, want empty host + http_status:404", r)
			}
		}
	}
}
```
