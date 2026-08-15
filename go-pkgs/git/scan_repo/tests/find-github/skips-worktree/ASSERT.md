## Expected

- Found path is `main`, not `feature-a`.

## Errors

- `err` is nil.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	wantMain := absPath(t, filepath.Join(req.Roots[0], "main"))
	if resp.Found == nil || resp.Found.Path != wantMain {
		t.Fatalf("Found.Path = %v, want %q", resp.Found, wantMain)
	}
}
```