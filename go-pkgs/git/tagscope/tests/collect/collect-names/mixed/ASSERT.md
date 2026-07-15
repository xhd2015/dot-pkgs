## Expected

- `resp.Collected.All` has length 2.
- `resp.Collected.Scopes` has length 2 (root and `sub/`).
- `resp.Collected.Unparsed` is `["release-1.0"]`.
- Parsed full names are `v0.0.1` and `sub/v0.0.2`.

## Errors

- `err` is nil.

```go
import (
	"slices"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	c := resp.Collected
	if len(c.All) != 2 {
		t.Fatalf("All len = %d, want 2", len(c.All))
	}
	if len(c.Scopes) != 2 {
		t.Fatalf("Scopes len = %d, want 2", len(c.Scopes))
	}
	if len(c.Unparsed) != 1 || c.Unparsed[0] != "release-1.0" {
		t.Fatalf("Unparsed = %v, want [release-1.0]", c.Unparsed)
	}
	names := []string{c.All[0].FullName, c.All[1].FullName}
	for _, want := range []string{"v0.0.1", "sub/v0.0.2"} {
		if !slices.Contains(names, want) {
			t.Fatalf("parsed names = %v, missing %q", names, want)
		}
	}
}
```