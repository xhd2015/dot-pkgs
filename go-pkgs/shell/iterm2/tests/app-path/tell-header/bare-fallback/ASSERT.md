## Expected

- Header is (or contains) `tell application "iTerm2"`.
- Header must not use `POSIX file` when appPath is empty.

## Exit Code

- N/A (library)

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	h := resp.Header
	if h == "" {
		t.Fatal("TellApplicationHeader returned empty")
	}
	want := bareTellTarget()
	if !hasBareTellTarget(h) {
		t.Fatalf("empty appPath must bare-fallback to %q; got %q", want, h)
	}
	if strings.Contains(h, "POSIX file") {
		t.Fatalf("empty appPath must not use POSIX file; got %q", h)
	}
}
```
