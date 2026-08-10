## Expected

- Tab `t1` Action is `skipped-busy` (or equivalent containing `skip` and `busy`).
- No Exec script contains `write text` for `busy-cmd`.
- `CreatedWindow` is false.

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

	if err != nil {
		t.Fatal(err)
	}
	var scripts []string
	cfg := buildRunConfig(req, &scripts)
	result, rerr := iterm2.RunTabSet(tabSetSpecFromReq(req), iterm2.TabSetRunOptions{
		Mode: iterm2.TabSetRunSmart,
	}, cfg)
	if rerr != nil {
		t.Fatalf("RunTabSet: %v", rerr)
	}
	if result == nil {
		t.Fatal("nil result")
	}
	if result.CreatedWindow {
		t.Fatal("CreatedWindow should be false when reusing")
	}
	act := actionForTab(result, "t1")
	if act != "skipped-busy" && !(strings.Contains(act, "skip") && strings.Contains(act, "busy")) {
		t.Fatalf("t1 Action = %q, want skipped-busy", act)
	}
	all := joinedScripts(scripts)
	if strings.Contains(all, `write text "busy-cmd"`) || strings.Contains(all, "busy-cmd") {
		// Allow busy-cmd only if not a write text resend — strict: no write text with command
		if strings.Contains(all, "write text") && strings.Contains(all, "busy-cmd") {
			t.Fatalf("busy tab must not resend write text busy-cmd; scripts:\n%s", all)
		}
	}
}
```
