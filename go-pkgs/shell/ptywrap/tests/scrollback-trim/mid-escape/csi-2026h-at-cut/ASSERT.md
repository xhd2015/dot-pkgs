## Expected

- `err` is nil.
- `resp.Trimmed` does **not** start with `026h` (orphan CSI tail).
- `resp.Trimmed` contains `TAIL_MARKER`.
- `resp.Trimmed` is non-empty and shorter than or equal to `req.Data`.

## Errors

- None.

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	got := resp.Trimmed
	if len(got) == 0 {
		t.Fatal("trimmed empty")
	}
	if len(got) > len(req.Data) {
		t.Fatalf("trimmed longer than input: %d > %d", len(got), len(req.Data))
	}
	if bytes.HasPrefix(got, []byte("026h")) {
		t.Fatalf("trimmed starts with orphan CSI tail 026h: head=%q", clip(got, 32))
	}
	// Also reject other short DEC tails from the same sequence if cut shifted.
	if bytes.HasPrefix(got, []byte("26h")) || bytes.HasPrefix(got, []byte("6h")) {
		t.Fatalf("trimmed starts with partial ?2026h tail: head=%q", clip(got, 32))
	}
	if !bytes.Contains(got, []byte("TAIL_MARKER")) {
		t.Fatalf("TAIL_MARKER missing after trim; head=%q", clip(got, 64))
	}
}

func clip(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n])
}
```
