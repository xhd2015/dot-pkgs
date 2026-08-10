## Expected

- Script embeds window title `Bots Window`.
- Script sets a window name (`set name` and `window` appear).

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
	if !strings.Contains(s, "Bots Window") {
		t.Fatalf("missing WindowName Bots Window; script:\n%s", s)
	}
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "set name") {
		t.Fatalf("script should set name for window; script:\n%s", s)
	}
	if !strings.Contains(lower, "window") {
		t.Fatalf("window name target should mention window; script:\n%s", s)
	}
}
```
