## Expected

- One decision for root scope.
- `LatestRelease` is `v0.0.2`.
- `NextTag` is `v0.0.3`.
- `SkipReason` is empty.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	dec := decisionFor(t, resp.Plan, "")
	if dec.LatestRelease != "v0.0.2" {
		t.Fatalf("LatestRelease = %q, want v0.0.2", dec.LatestRelease)
	}
	if dec.NextTag != "v0.0.3" {
		t.Fatalf("NextTag = %q, want v0.0.3", dec.NextTag)
	}
	if dec.SkipReason != "" {
		t.Fatalf("SkipReason = %q, want empty", dec.SkipReason)
	}
}
```