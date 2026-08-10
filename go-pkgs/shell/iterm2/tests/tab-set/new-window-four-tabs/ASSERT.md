## Expected

- Non-empty AppleScript.
- Exactly one `create window with default profile`.
- Exactly three `create tab with default profile` (N−1 for N=4; first tab is the
  window’s initial session).
- `write text` lines for `cmd-a`, `cmd-b`, `cmd-c`, `cmd-d` appear in that order.

## Exit Code

- N/A (build-tab-set-script phase)

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
	s := resp.Script
	if s == "" {
		t.Fatal("expected non-empty BuildTabSetNewWindowScript output")
	}
	const createWindow = `create window with default profile`
	const createTab = `create tab with default profile`
	if n := countOccurrences(s, createWindow); n != 1 {
		t.Fatalf("create window count = %d, want 1; script:\n%s", n, s)
	}
	if n := countOccurrences(s, createTab); n != 3 {
		t.Fatalf("create tab count = %d, want 3 (N-1); script:\n%s", n, s)
	}
	cmds := []string{"cmd-a", "cmd-b", "cmd-c", "cmd-d"}
	prev := -1
	for _, cmd := range cmds {
		line := writeTextLine(cmd)
		idx := strings.Index(s, line)
		if idx < 0 {
			// allow EscapeCommandForAppleScript form if already escaped
			idx = strings.Index(s, `write text "`+cmd)
		}
		if idx < 0 {
			t.Fatalf("missing write text for %q; script:\n%s", cmd, s)
		}
		if idx < prev {
			t.Fatalf("command order wrong: %q appears before previous; script:\n%s", cmd, s)
		}
		prev = idx
	}
}
```
