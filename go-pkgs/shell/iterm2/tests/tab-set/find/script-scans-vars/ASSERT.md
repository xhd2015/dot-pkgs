## Expected

- `BuildTabSetFindScript(setName)` returns non-empty AppleScript.
- Script mentions both marker variables:
  - `user.koolTabSet`
  - `user.koolTabSetTab`
- Script includes the set name (`bots`) for filter or emission.
- Script scans sessions (mentions `session` / windows-tabs structure).

## Exit Code

- N/A (build-find-script phase)

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	script := iterm2.BuildTabSetFindScript(req.TabSetName)
	if script == "" {
		t.Fatal("BuildTabSetFindScript returned empty script")
	}
	setVar := `variable named "` + tabSetVarLiteral + `"`
	tabVar := `variable named "` + tabSetTabVarLiteral + `"`
	// Also accept bare variable name strings without "variable named"
	if !strings.Contains(script, setVar) && !strings.Contains(script, tabSetVarLiteral) {
		t.Fatalf("find script must scan %s; script:\n%s", tabSetVarLiteral, script)
	}
	if !strings.Contains(script, tabVar) && !strings.Contains(script, tabSetTabVarLiteral) {
		t.Fatalf("find script must scan %s; script:\n%s", tabSetTabVarLiteral, script)
	}
	if !strings.Contains(script, req.TabSetName) {
		t.Fatalf("find script must include set name %q; script:\n%s", req.TabSetName, script)
	}
	lower := strings.ToLower(script)
	if !strings.Contains(lower, "session") {
		t.Fatalf("find script should scan sessions; script:\n%s", script)
	}
	// Inside tell application "iTerm2", bare `tab` is an iTerm element, not a
	// field delimiter. Must use ASCII character 9 (or equivalent), not bare tab.
	if !strings.Contains(script, "ASCII character 9") && !strings.Contains(script, "fieldSep") {
		t.Fatalf("find script must use ASCII TAB field separator (not bare AppleScript tab); script:\n%s", script)
	}
	// Reject the historical bug: joining with bare ` & tab & ` in iTerm tell.
	if strings.Contains(script, " & tab & ") {
		t.Fatalf("find script must not use bare AppleScript tab as delimiter inside iTerm tell; script:\n%s", script)
	}
}
```
