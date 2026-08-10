## Expected

- Non-empty force-new script.
- Embeds path-bound tell for AppPath (same as `TellApplicationHeader(AppPath)`).
- Does not use bare `tell application "iTerm2"`.
- Still creates a new window (force-new behavior retained).

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
	if resp.Header != "" && !strings.Contains(resp.Script, strings.TrimSpace(resp.Header)) {
		// Header from TellApplicationHeader should appear in script when both use helper.
		if !hasPathBoundTell(resp.Script, req.AppPath) {
			t.Fatalf("script should embed header %q; script:\n%s", resp.Header, resp.Script)
		}
	}
	if !strings.Contains(resp.Script, "create window with default profile") {
		t.Fatalf("force-new must still create window; script:\n%s", resp.Script)
	}
}
```
