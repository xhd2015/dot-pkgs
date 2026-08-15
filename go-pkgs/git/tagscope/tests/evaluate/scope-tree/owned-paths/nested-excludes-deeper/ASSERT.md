## Expected

- `OwnedPaths` is `["sub/a.txt", "sub/pkg.go"]` (excludes `sub/nested/` paths).

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
	want := []string{"sub/a.txt", "sub/pkg.go"}
	if len(resp.OwnedPaths) != len(want) {
		t.Fatalf("OwnedPaths = %v, want %v", resp.OwnedPaths, want)
	}
	for i := range want {
		if resp.OwnedPaths[i] != want[i] {
			t.Fatalf("OwnedPaths = %v, want %v", resp.OwnedPaths, want)
		}
	}
}
```