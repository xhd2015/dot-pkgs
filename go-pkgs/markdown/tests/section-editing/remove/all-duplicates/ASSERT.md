## Expected

Both matching sections, including the nested child of the first, are removed.
The unrelated user section remains byte-for-byte unchanged.

## Errors

- The operation returns no error.
- The removed count is exactly two.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp.Removed != 2 {
		t.Fatalf("removed = %d, want 2", resp.Removed)
	}
	assertExact(t, resp.Output, "# User\nkeep\n")
	assertExact(t, resp.SecondOutput, "# User\nkeep\n")
}
```
