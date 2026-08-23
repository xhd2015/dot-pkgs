## Expected

- `resp.Formatted` is `"dirty (0 staged, 1 changed, 0 renamed, 0 deleted, 0 untracked)"`.

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
	want := "dirty (0 staged, 1 changed, 0 renamed, 0 deleted, 0 untracked)"
	if resp.Formatted != want {
		t.Fatalf("formatted = %q, want %q", resp.Formatted, want)
	}
}
```
