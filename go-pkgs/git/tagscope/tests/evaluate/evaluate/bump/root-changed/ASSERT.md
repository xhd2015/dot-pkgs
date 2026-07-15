## Expected

- One decision for root scope.
- `LatestRelease` is `v0.0.2`.
- `NextTag` is `v0.0.3`.
- `SkipReason` is empty.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	d := decisionFor(t, resp.Plan, "")
	if d.LatestRelease != "v0.0.2" {
		t.Fatalf("LatestRelease = %q, want v0.0.2", d.LatestRelease)
	}
	if d.NextTag != "v0.0.3" {
		t.Fatalf("NextTag = %q, want v0.0.3", d.NextTag)
	}
	if d.SkipReason != "" {
		t.Fatalf("SkipReason = %q, want empty", d.SkipReason)
	}
}
```