## Expected

- `err` is nil.
- `resp.IsTTY` is false for a `bytes.Buffer`.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- N/A (in-process library).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("WriterIsTTY(bytes.Buffer): unexpected error: %v", err)
	}
	if resp.IsTTY {
		t.Fatal("WriterIsTTY(bytes.Buffer): IsTTY=true, want false")
	}
}
```
