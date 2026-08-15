## Expected

- `err` is nil.
- `resp.Replies` is empty (no CPR written).
- `resp.Rest` is nil or empty.
- `resp.WriteCalls` is 0.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Replies) != 0 {
		t.Fatalf("disabled should not write CPR, got %q", resp.Replies)
	}
	if len(resp.Rest) != 0 {
		t.Fatalf("disabled should return nil/empty rest, got %q", resp.Rest)
	}
	if resp.WriteCalls != 0 {
		t.Fatalf("disabled WriteCalls=%d want 0", resp.WriteCalls)
	}
}
```
