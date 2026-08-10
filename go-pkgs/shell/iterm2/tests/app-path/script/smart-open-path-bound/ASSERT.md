## Expected

- Non-empty smart-open script.
- Path-bound tell for AppPath; no bare `"iTerm2"` target.
- Retains smart-open markers (path scan / create tab or window) so body is still smart-open.

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
	assertPathBoundScript(t, resp.Script, req.AppPath)
	s := resp.Script
	// Representative smart-open body checks (not full branch suite).
	if !strings.Contains(s, "create window with default profile") &&
		!strings.Contains(s, "create tab with default profile") {
		t.Fatalf("smart-open script should create tab and/or window; script:\n%s", s)
	}
	if !strings.Contains(s, "targetDir") {
		t.Fatalf("smart-open script should set targetDir; script:\n%s", s)
	}
}
```
