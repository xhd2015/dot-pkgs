## Expected

- Tab `t1` Action is `resent`.
- Some Exec script contains `write text` with `idle-cmd`.
- No Ctrl+C / control-c keystroke in any Exec script.

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
	act := actionForTab(result, "t1")
	if act != "resent" {
		t.Fatalf("t1 Action = %q, want resent", act)
	}
	all := joinedScripts(scripts)
	if !strings.Contains(all, "idle-cmd") || !strings.Contains(all, "write text") {
		t.Fatalf("expected write text resend of idle-cmd; scripts:\n%s", all)
	}
	if scriptsContainCtrlC(scripts) {
		t.Fatalf("resend must not send Ctrl+C; scripts:\n%s", all)
	}
}
```
