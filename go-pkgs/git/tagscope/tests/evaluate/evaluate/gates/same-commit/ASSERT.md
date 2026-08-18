## Expected

- One decision for root scope.
- `LatestRelease` is `v0.0.2`.
- `NextTag` is empty.
- `SkipReason` is `same-commit`.

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
	if len(resp.Plan.Decisions) != 1 {
		t.Fatalf("Decisions len = %d, want 1", len(resp.Plan.Decisions))
	}
	dec := resp.Plan.Decisions[0]
	if dec.LatestRelease != "v0.0.2" {
		t.Fatalf("LatestRelease = %q, want v0.0.2", dec.LatestRelease)
	}
	if dec.NextTag != "" {
		t.Fatalf("NextTag = %q, want empty", dec.NextTag)
	}
	if dec.SkipReason != "same-commit" {
		t.Fatalf("SkipReason = %q, want same-commit", dec.SkipReason)
	}
}
```