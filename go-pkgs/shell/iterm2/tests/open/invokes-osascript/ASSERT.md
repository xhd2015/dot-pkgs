## Expected

- `err` is nil; captured script is non-empty and contains absolute directory.

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
	if resp.CapturedScript == "" {
		t.Fatal("expected osascript to be called")
	}
	abs, err := filepath.Abs(req.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.CapturedScript, abs) {
		t.Fatalf("script missing dir %q: %q", abs, resp.CapturedScript)
	}
}
```