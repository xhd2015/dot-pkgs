## Expected

- Non-empty AppleScript.
- Contains `activate`.
- Contains window id `win-focus-42`.
- Mentions the stable session ID and records its containing tab before selection.
- Contains the fallback tab index `2`.
- Does **not** create a window (`create window` absent).

## Exit Code

- N/A (library)

```go
import (
	"strconv"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	script := resp.Script
	if script == "" {
		t.Fatal("BuildFocusScript returned empty script")
	}
	lower := strings.ToLower(script)
	if !strings.Contains(lower, "activate") {
		t.Fatalf("focus script must activate iTerm; script:\n%s", script)
	}
	if !strings.Contains(script, req.FocusRef.WindowID) {
		t.Fatalf("focus script must include window id %q; script:\n%s", req.FocusRef.WindowID, script)
	}
	if !strings.Contains(script, req.FocusRef.SessionID) {
		t.Fatalf("focus script must include stable session id %q; script:\n%s", req.FocusRef.SessionID, script)
	}
	if !strings.Contains(script, "set targetTab to aTab") || !strings.Contains(script, "select targetTab") {
		t.Fatalf("focus script must select the tab containing the stable session; script:\n%s", script)
	}
	// Keep the numeric tab index as the fallback when no session ID is available.
	tabStr := strconv.Itoa(req.FocusRef.TabIndex)
	if !strings.Contains(script, tabStr) {
		t.Fatalf("focus script must include tab index %s; script:\n%s", tabStr, script)
	}
	if !strings.Contains(lower, "tab") {
		t.Fatalf("focus script must select a tab; script:\n%s", script)
	}
	// Select window somehow (select / id of aWindow / targetWindow patterns).
	if !strings.Contains(lower, "select") && !strings.Contains(lower, "window") {
		t.Fatalf("focus script must select window; script:\n%s", script)
	}
	if strings.Contains(lower, "create window") {
		t.Fatalf("focus script must not create a window; script:\n%s", script)
	}
}
```
