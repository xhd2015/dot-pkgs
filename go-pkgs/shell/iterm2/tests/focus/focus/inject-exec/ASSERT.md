## Expected

- `Focus(ref, cfg)` returns nil.
- `cfg.Exec` invoked at least once.
- Some Exec script contains the window id and tab index from `FocusRef`.
- Exec script looks like a focus script (contains `activate`).

## Exit Code

- N/A (library)

```go
import (
	"strconv"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	var scripts []string
	cfg := &iterm2.FocusConfig{
		Exec: func(script string) error {
			scripts = append(scripts, script)
			return nil
		},
	}
	ferr := iterm2.Focus(sessionRefFromInput(req.FocusRef), cfg)
	if ferr != nil {
		t.Fatalf("Focus: %v", ferr)
	}
	if len(scripts) == 0 {
		t.Fatal("Focus must call Exec at least once")
	}
	all := strings.Join(scripts, "\n---\n")
	if !strings.Contains(all, req.FocusRef.WindowID) {
		t.Fatalf("Exec script must contain window id %q; scripts:\n%s", req.FocusRef.WindowID, all)
	}
	tabStr := strconv.Itoa(req.FocusRef.TabIndex)
	if !strings.Contains(all, tabStr) {
		t.Fatalf("Exec script must contain tab index %s; scripts:\n%s", tabStr, all)
	}
	if !strings.Contains(strings.ToLower(all), "activate") {
		t.Fatalf("Exec script should activate iTerm; scripts:\n%s", all)
	}
}
```
