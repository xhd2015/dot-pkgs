## Expected

- Each kool `go1.Y` input pins to the listed patch version.
- Naked `1.19` pins to `go1.19.13` (same as `go1.19`).

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
	cases := koolPinCases()
	if len(resp.Pins) != len(cases) {
		t.Fatalf("PinPatch table size = %d, want %d (%v)", len(resp.Pins), len(cases), resp.Pins)
	}
	for _, c := range cases {
		got := resp.Pins[c[0]]
		if got != c[1] {
			t.Fatalf("PinPatch(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}
```
