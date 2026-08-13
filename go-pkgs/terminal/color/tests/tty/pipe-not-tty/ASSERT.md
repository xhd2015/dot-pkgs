## Expected

- `err` is nil.
- `resp.IsTTY` is false for an `os.Pipe` writer.

## Side Effects

- None from the product. The harness opens and closes a pipe pair.

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
		t.Fatalf("WriterIsTTY(os.Pipe writer): unexpected error: %v", err)
	}
	if resp.IsTTY {
		t.Fatal("WriterIsTTY(os.Pipe writer): IsTTY=true, want false")
	}
}
```
