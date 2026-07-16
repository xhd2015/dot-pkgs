## Expected

- Attach succeeds.
- Fake runner recorded at least one Exec whose args contain `route` and `dns`.
- That same call (or another route dns call) includes hostname `a.example.com`.
- RouteDNSCount ≥ 1.

## Side Effects

- DNS is faked only — no real Cloudflare API.

## Errors

- None.

## Exit Code

- 0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Attach error: %v", err)
	}
	if resp == nil || resp.Runner == nil {
		t.Fatal("nil response or Runner")
	}
	const host = "a.example.com"
	if resp.RouteDNSCount < 1 {
		t.Fatalf("RouteDNSCount = %d, want ≥ 1", resp.RouteDNSCount)
	}
	if !resp.Runner.hasRouteDNSFor(host) {
		t.Fatalf("fake runner missing route dns call for %q; calls=%v", host, resp.Runner.calls)
	}
}
```
