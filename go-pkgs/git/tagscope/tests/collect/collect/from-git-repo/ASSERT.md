## Expected

- `resp.Collected.All` has length 3 (parsed tags only).
- `resp.Collected.Unparsed` contains `release-1.0`.
- Root scope `Newest.FullName` is `v0.0.2`.
- `sub/` scope `Newest.FullName` is `sub/v0.0.1`.

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
	if len(c.All) != 3 {
		t.Fatalf("All len = %d, want 3", len(c.All))
	}
	if len(c.Unparsed) != 1 || c.Unparsed[0] != "release-1.0" {
		t.Fatalf("Unparsed = %v, want [release-1.0]", c.Unparsed)
	}
	names := make([]string, len(c.All))
	for i, tag := range c.All {
		names[i] = tag.FullName
	}
	for _, want := range []string{"v0.0.1", "v0.0.2", "sub/v0.0.1"} {
		if !slices.Contains(names, want) {
			t.Fatalf("parsed names = %v, missing %q", names, want)
		}
	}
	root := lineageFor(t, c, "")
	sub := lineageFor(t, c, "sub/")
	if root.Newest == nil || root.Newest.FullName != "v0.0.2" {
		t.Fatalf("root Newest = %v, want v0.0.2", root.Newest)
	}
	if sub.Newest == nil || sub.Newest.FullName != "sub/v0.0.1" {
		t.Fatalf("sub Newest = %v, want sub/v0.0.1", sub.Newest)
	}
}
```