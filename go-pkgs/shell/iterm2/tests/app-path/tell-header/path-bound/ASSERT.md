## Expected

- Header contains path-bound `tell application "<escaped appPath>"` form.
- Header does **not** use bare `tell application "iTerm2"` as the tell target.
- Header does **not** use `POSIX file` expression form (breaks iTerm dictionary load).

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
	want := pathBoundTellLine(req.AppPath)
	if !hasPathBoundTell(h, req.AppPath) {
		t.Fatalf("header must be path-bound; want like %q; got %q", want, h)
	}
	if hasBareTellTarget(h) {
		t.Fatalf("path-bound header must not use bare target; got %q", h)
	}
	if strings.Contains(h, "POSIX file") {
		t.Fatalf("path-bound header must not use POSIX file expression; got %q", h)
	}
	// Prefer exact line match when product returns a single-line header.
	if strings.TrimSpace(h) != want && !strings.Contains(h, want) {
		t.Fatalf("header %q should equal or contain %q", h, want)
	}
}
```
