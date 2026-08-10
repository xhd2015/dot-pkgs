## Expected

- Non-empty smoke script.
- Path-bound tell for AppPath; no bare `"iTerm2"` target.
- Still probes session `path` / returns ok-style smoke structure.

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
	if !strings.Contains(s, `variable named "path"`) && !strings.Contains(s, `named "path"`) {
		t.Fatalf("smoke script should probe session path variable; script:\n%s", s)
	}
}
```
