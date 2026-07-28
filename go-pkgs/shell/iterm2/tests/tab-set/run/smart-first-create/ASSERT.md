## Expected

- `RunTabSet` succeeds with `CreatedWindow == true`.
- At least one Exec script contains `create window with default profile`.
- Exec scripts include both tab commands `cmd-a` and `cmd-b`.

## Exit Code

- N/A (run-tab-set)

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
		Mode: runModeFromReq(req),
	}, cfg)
	if rerr != nil {
		t.Fatalf("RunTabSet: %v", rerr)
	}
	if result == nil {
		t.Fatal("RunTabSet returned nil result")
	}
	if !result.CreatedWindow {
		t.Fatalf("CreatedWindow = false, want true; result=%+v", result)
	}
	all := joinedScripts(scripts)
	if !strings.Contains(all, `create window with default profile`) {
		t.Fatalf("Exec scripts missing create window; scripts:\n%s", all)
	}
	if !strings.Contains(all, "cmd-a") || !strings.Contains(all, "cmd-b") {
		t.Fatalf("Exec scripts missing tab commands; scripts:\n%s", all)
	}
}
```
