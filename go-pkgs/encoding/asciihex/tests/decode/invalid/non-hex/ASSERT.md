## Expected

- `err` is non-nil.
- `err.Error()` equals `invalid hex value GG: ` plus the `strconv.ParseInt("GG", 16, 32)` error (kool `decodeAsciiHex`).

## Side Effects

- None.

## Errors

- Format: `invalid hex value GG: %v` where `%v` is `strconv.ParseInt("GG", 16, 32)`.

## Exit Code

- N/A (in-process library).

```go
import (
	"fmt"
	"strconv"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("Decode(%q): want error, got nil", req.Hex)
	}
	_, parseErr := strconv.ParseInt("GG", 16, 32)
	if parseErr == nil {
		t.Fatal("strconv.ParseInt(\"GG\", 16, 32) unexpectedly succeeded")
	}
	want := fmt.Sprintf("invalid hex value %s: %v", "GG", parseErr)
	if got := err.Error(); got != want {
		t.Fatalf("Decode(%q): error %q, want %q", req.Hex, got, want)
	}
}
```
