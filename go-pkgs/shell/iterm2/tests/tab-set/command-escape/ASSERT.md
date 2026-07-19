## Expected

- Script contains a `write text` line with the command escaped the same way as
  `EscapeCommandForAppleScript`: `\` → `\\`, `"` → `\"`.
- Expected embedded form: `echo \"hi\"\\x` inside a double-quoted AppleScript string.

## Exit Code

- N/A (build-tab-set-script phase)

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
	s := resp.Script
	if s == "" {
		t.Fatal("expected non-empty script")
	}
	raw := `echo "hi"\x`
	escaped := iterm2.EscapeCommandForAppleScript(raw)
	wantLine := `write text "` + escaped + `"`
	if !strings.Contains(s, wantLine) {
		// Also accept if escape landed correctly even with spacing variants
		if !strings.Contains(s, escaped) {
			t.Fatalf("missing escaped command %q (want line %q); script:\n%s", escaped, wantLine, s)
		}
		t.Fatalf("escaped form present but not as write text line %q; script:\n%s", wantLine, s)
	}
	// Unescaped quote form must not appear as a raw write text payload
	if strings.Contains(s, `write text "echo "hi"`) {
		t.Fatalf("command quotes not escaped in write text; script:\n%s", s)
	}
}
```
