## Expected

- `resp.Display` is `"child"` (platform-native separators).

## Errors

- `err` is nil.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	want := "child"
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (cwd=%q path=%q)", want, resp.Display, resp.Cwd, req.Path)
	}
}```
