## Expected

- `RunTabSet` with `TabSetRunNewWindow` Exec's a create-window script.
- Result has `CreatedWindow == true` (or equivalent always-create signal).
- Create path runs despite non-empty Find (force new).

## Exit Code

- N/A

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d

	if err != nil {
		t.Fatal(err)
	}
	var scripts []string
	cfg := buildRunConfig(req, &scripts)
	result, rerr := iterm2.RunTabSet(tabSetSpecFromReq(req), iterm2.TabSetRunOptions{
		Mode: iterm2.TabSetRunNewWindow,
	}, cfg)
	if rerr != nil {
		t.Fatalf("RunTabSet: %v", rerr)
	}
	if result == nil {
		t.Fatal("nil result")
	}
	if !result.CreatedWindow {
		t.Fatalf("NewWindow mode: CreatedWindow = false, want true")
	}
	all := joinedScripts(scripts)
	if !strings.Contains(all, `create window with default profile`) {
		t.Fatalf("NewWindow must Exec create window; scripts:\n%s", all)
	}
	if !strings.Contains(all, "force-new-cmd") {
		t.Fatalf("missing force-new-cmd in create scripts:\n%s", all)
	}
}
```
