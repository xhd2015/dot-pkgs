## Expected

- `err` is nil.
- `resp.Encoded` equals exactly `\x00\x7f\xff` (two digits, lowercase, including high byte `ff`).
- `resp.Encoded` does not contain a newline.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- N/A (in-process library).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Encode(%q): unexpected error: %v", req.Data, err)
	}
	const want = `\x00\x7f\xff`
	if resp.Encoded != want {
		t.Fatalf("Encode(%q): got %q, want %q", req.Data, resp.Encoded, want)
	}
	if strings.Contains(resp.Encoded, "\n") {
		t.Fatalf("Encode(%q): output contains newline: %q", req.Data, resp.Encoded)
	}
}
```
