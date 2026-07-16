## Expected

- Final `err == nil` and last Stop error nil.
- Two Attach sessions recorded; two Stop steps applied.
- State Hosts empty.
- `ConnectorPID == 0`.
- Config has no hostname rules (404-only OK).

## Side Effects

- After intermediate Stop(A) siblings-remain is assumed (covered by sibling leaf);
  this leaf asserts **final** empty teardown only.

## Errors

- None.

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Attach/Stop sequence error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.LastStopErr != nil {
		t.Fatalf("last Stop error: %v", resp.LastStopErr)
	}
	if len(resp.Sessions) != 2 {
		t.Fatalf("Sessions len = %d, want 2", len(resp.Sessions))
	}
	if resp.State == nil {
		t.Fatal("nil State after Stop(A)+Stop(B)")
	}
	if len(resp.State.Hosts) != 0 {
		t.Fatalf("Hosts len = %d, want 0; Hosts=%v", len(resp.State.Hosts), resp.State.Hosts)
	}
	if resp.State.ConnectorPID != 0 {
		t.Fatalf("ConnectorPID = %d, want 0 after last host gone", resp.State.ConnectorPID)
	}

	if resp.Config != nil {
		svcs := hostRuleServices(resp.Config)
		if len(svcs) != 0 {
			t.Fatalf("host rules remain: %v", svcs)
		}
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
