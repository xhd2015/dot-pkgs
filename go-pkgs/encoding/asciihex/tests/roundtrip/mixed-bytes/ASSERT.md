## Expected

- `err` is nil.
- `resp.Decoded` equals `req.Data` byte-for-byte.
- `0xff` comes back as the single byte `0xff`, not UTF-8 `c3 bf` (U+00FF via `WriteRune`).
- `resp.Encoded` has no newline.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- N/A (in-process library).

```go
import (
	"bytes"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Decode(Encode(%q)): unexpected error: %v", req.Data, err)
	}
	if strings.Contains(resp.Encoded, "\n") {
		t.Fatalf("Encode(%q): output contains newline: %q", req.Data, resp.Encoded)
	}
	if !bytes.Equal(resp.Decoded, req.Data) {
		t.Fatalf("Decode(Encode(%q)): got %q (%x), want %q (%x) — 0xff must be one raw byte, not UTF-8 U+00FF",
			req.Data, resp.Decoded, resp.Decoded, req.Data, req.Data)
	}
}
```
