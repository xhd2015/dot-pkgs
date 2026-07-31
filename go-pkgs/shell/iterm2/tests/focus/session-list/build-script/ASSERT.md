## Expected

- Non-empty AppleScript.
- Mentions windows, tabs, and sessions structure.
- Mentions `tty` (session tty property).
- Uses safe TAB field separator (`ASCII character 9` and/or `fieldSep`).
- Does **not** join fields with bare ` & tab & ` inside the iTerm tell block.

## Exit Code

- N/A (library)

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	script := resp.Script
	if script == "" {
		t.Fatal("BuildSessionListScript returned empty script")
	}
	lower := strings.ToLower(script)
	for _, need := range []string{"window", "tab", "session", "tty"} {
		if !strings.Contains(lower, need) {
			t.Fatalf("list script should mention %q; script:\n%s", need, script)
		}
	}
	if !strings.Contains(script, "ASCII character 9") && !strings.Contains(script, "fieldSep") {
		t.Fatalf("list script must use ASCII TAB field separator (not bare AppleScript tab); script:\n%s", script)
	}
	if strings.Contains(script, " & tab & ") {
		t.Fatalf("list script must not use bare AppleScript tab as delimiter inside iTerm tell; script:\n%s", script)
	}
}
```
