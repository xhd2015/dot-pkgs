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
	d := resp.Plan.Decisions[0]
	if d.Scope.PathPrefix != "" {
		t.Fatalf("Scope.PathPrefix = %q, want root", d.Scope.PathPrefix)
	}
	if d.LatestRelease != "" {
		t.Fatalf("LatestRelease = %q, want empty", d.LatestRelease)
	}
	if d.NextTag != "" {
		t.Fatalf("NextTag = %q, want empty", d.NextTag)
	}
	if d.SkipReason != "no-baseline" {
		t.Fatalf("SkipReason = %q, want no-baseline", d.SkipReason)
	}
}
```