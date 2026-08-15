## Expected

- `resp.Collected.Scopes` has length 3.
- Scope `VersionPrefix` order is `a/v`, `v`, `z/v`.
- Scope `PathPrefix` order is `a/`, `""`, `z/`.

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
	scopes := resp.Collected.Scopes
	if len(scopes) != 3 {
		t.Fatalf("Scopes len = %d, want 3", len(scopes))
	}
	wantPrefixes := []string{"a/v", "v", "z/v"}
	wantPaths := []string{"a/", "", "z/"}
	for i := range wantPrefixes {
		if scopes[i].VersionPrefix != wantPrefixes[i] {
			t.Fatalf("Scopes[%d].VersionPrefix = %q, want %q", i, scopes[i].VersionPrefix, wantPrefixes[i])
		}
		if scopes[i].PathPrefix != wantPaths[i] {
			t.Fatalf("Scopes[%d].PathPrefix = %q, want %q", i, scopes[i].PathPrefix, wantPaths[i])
		}
	}
}
```