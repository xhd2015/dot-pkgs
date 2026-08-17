## Expected

- `PinPatch("go1.19.13")` is `go1.19.13`.
- `PinPatch("1.19.13")` is `go1.19.13` (prefix added, patch unchanged).

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	want := map[string]string{
		"go1.19.13": "go1.19.13",
		"1.19.13":   "go1.19.13",
	}
	if len(resp.Pins) != len(want) {
		t.Fatalf("PinPatch table size = %d, want %d (%v)", len(resp.Pins), len(want), resp.Pins)
	}
	for in, exp := range want {
		if got := resp.Pins[in]; got != exp {
			t.Fatalf("PinPatch(%q) = %q, want %q", in, got, exp)
		}
	}
}
```
