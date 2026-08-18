## Expected

- One decision for `sub/` scope.
- `LatestRelease` is `sub/v0.2.3`.
- `NextTag` is `sub/v0.2.4`.
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
	dec := decisionFor(t, resp.Plan, "sub/")
	if dec.LatestRelease != "sub/v0.2.3" {
		t.Fatalf("LatestRelease = %q, want sub/v0.2.3", dec.LatestRelease)
	}
	if dec.NextTag != "sub/v0.2.4" {
		t.Fatalf("NextTag = %q, want sub/v0.2.4", dec.NextTag)
	}
	if dec.SkipReason != "" {
		t.Fatalf("SkipReason = %q, want empty", dec.SkipReason)
	}
}
```