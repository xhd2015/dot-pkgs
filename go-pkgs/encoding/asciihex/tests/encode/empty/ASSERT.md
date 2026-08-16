## Expected

- `err` is nil.
- `resp.Encoded` is `""`.

## Side Effects

- None. Encode is a pure function of its argument.

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
		t.Fatalf("Encode([]byte{}): unexpected error: %v", err)
	}
	if resp.Encoded != "" {
		t.Fatalf("Encode([]byte{}): got %q, want \"\"", resp.Encoded)
	}
}
```
