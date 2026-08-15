## Expected

- `resp.Formatted` is `"dirty (1 added, 1 changed, 1 renamed, 1 deleted)"`.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := "dirty (1 added, 1 changed, 1 renamed, 1 deleted)"
	if resp.Formatted != want {
		t.Fatalf("formatted = %q, want %q", resp.Formatted, want)
	}
}
```