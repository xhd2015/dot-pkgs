## Expected

- One decision for root scope.
- `LatestRelease` is `v0.0.2`.
- `NextTag` is empty.
- `SkipReason` is `prerelease-head`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Plan.Decisions) != 1 {
		t.Fatalf("Decisions len = %d, want 1", len(resp.Plan.Decisions))
	}
	d := resp.Plan.Decisions[0]
	if d.LatestRelease != "v0.0.2" {
		t.Fatalf("LatestRelease = %q, want v0.0.2", d.LatestRelease)
	}
	if d.NextTag != "" {
		t.Fatalf("NextTag = %q, want empty", d.NextTag)
	}
	if d.SkipReason != "prerelease-head" {
		t.Fatalf("SkipReason = %q, want prerelease-head", d.SkipReason)
	}
}
```