## Expected

- `resp.Collected.All` is empty.
- `resp.Collected.Scopes` is empty.
- `resp.Collected.ByScope` is empty.
- `resp.Collected.Unparsed` contains `release-1.0` and `v0.0`.

## Errors

- `err` is nil.

```go
import (
	"slices"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	c := resp.Collected
	if len(c.All) != 0 || len(c.Scopes) != 0 || len(c.ByScope) != 0 {
		t.Fatalf("expected no parsed inventory, got All=%d Scopes=%d ByScope=%d", len(c.All), len(c.Scopes), len(c.ByScope))
	}
	want := []string{"release-1.0", "v0.0"}
	if len(c.Unparsed) != len(want) {
		t.Fatalf("Unparsed = %v, want %v", c.Unparsed, want)
	}
	for _, name := range want {
		if !slices.Contains(c.Unparsed, name) {
			t.Fatalf("Unparsed = %v, missing %q", c.Unparsed, name)
		}
	}
}
```