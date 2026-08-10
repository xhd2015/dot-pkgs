## Expected

- Exactly one `create window with default profile`.
- Zero `create tab with default profile` (initial session is the only tab).
- Command `solo-cmd` appears as a `write text` line.

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
		t.Fatal("expected non-empty script")
	}
	const createWindow = `create window with default profile`
	const createTab = `create tab with default profile`
	if n := countOccurrences(s, createWindow); n != 1 {
		t.Fatalf("create window count = %d, want 1; script:\n%s", n, s)
	}
	if n := countOccurrences(s, createTab); n != 0 {
		t.Fatalf("create tab count = %d, want 0 for single-tab set; script:\n%s", n, s)
	}
	if !strings.Contains(s, writeTextLine("solo-cmd")) && !strings.Contains(s, `write text "solo-cmd`) {
		t.Fatalf("missing write text for solo-cmd; script:\n%s", s)
	}
}
```
