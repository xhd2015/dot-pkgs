## Expected

- Captured script contains canonical real path.

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	real, err := filepath.EvalSymlinks(req.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.CapturedScript, real) {
		t.Fatalf("script %q missing canonical %q", resp.CapturedScript, real)
	}
}
```