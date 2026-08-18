## Expected

- One decision for root scope.
- `LatestRelease` field is empty.
- `NextTag` is empty.
- `SkipReason` is `no-baseline`.

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
	if resp.Plan.Head != req.HeadCommit {
		t.Fatalf("Plan.Head = %q, want %q", resp.Plan.Head, req.HeadCommit)
	}
	if len(resp.Plan.Decisions) != 1 {
		t.Fatalf("Decisions len = %d, want 1", len(resp.Plan.Decisions))
	}
	dec := resp.Plan.Decisions[0]
	if dec.Scope.PathPrefix != "" {
		t.Fatalf("Scope.PathPrefix = %q, want root", dec.Scope.PathPrefix)
	}
	if dec.LatestRelease != "" {
		t.Fatalf("LatestRelease = %q, want empty", dec.LatestRelease)
	}
	if dec.NextTag != "" {
		t.Fatalf("NextTag = %q, want empty", dec.NextTag)
	}
	if dec.SkipReason != "no-baseline" {
		t.Fatalf("SkipReason = %q, want no-baseline", dec.SkipReason)
	}
}
```