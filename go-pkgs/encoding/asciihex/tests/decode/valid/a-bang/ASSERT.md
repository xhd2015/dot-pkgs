## Expected

- `err` is nil.
- `resp.Decoded` equals `[]byte("A!")` (`0x41`, `0x21`).

## Side Effects

- None.

## Errors

- None.

## Exit Code

- N/A (in-process library).

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Decode(%q): unexpected error: %v", req.Hex, err)
	}
	want := []byte("A!")
	if !bytes.Equal(resp.Decoded, want) {
		t.Fatalf("Decode(%q): got %q, want %q", req.Hex, resp.Decoded, want)
	}
}
```
