## Expected

- Captured script contains canonical real path.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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