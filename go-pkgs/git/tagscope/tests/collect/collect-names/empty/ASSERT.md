## Expected

- `resp.Collected.All` is empty.
- `resp.Collected.Scopes` is empty.
- `resp.Collected.ByScope` is empty.
- `resp.Collected.Unparsed` is empty.

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
	c := resp.Collected
	if len(c.All) != 0 {
		t.Fatalf("All len = %d, want 0", len(c.All))
	}
	if len(c.Scopes) != 0 {
		t.Fatalf("Scopes len = %d, want 0", len(c.Scopes))
	}
	if len(c.ByScope) != 0 {
		t.Fatalf("ByScope len = %d, want 0", len(c.ByScope))
	}
	if len(c.Unparsed) != 0 {
		t.Fatalf("Unparsed len = %d, want 0", len(c.Unparsed))
	}
}
```