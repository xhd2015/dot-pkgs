## Expected

- Run succeeds (idle resend).
- No Exec script contains Ctrl+C automation (`control-c`, `ctrl-c`,
  `keystroke "c" using control`, etc.).

## Exit Code

- N/A

```go
import (
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
	_, rerr := iterm2.RunTabSet(tabSetSpecFromReq(req), iterm2.TabSetRunOptions{
		Mode: iterm2.TabSetRunSmart,
	}, cfg)
	if rerr != nil {
		t.Fatalf("RunTabSet: %v", rerr)
	}
	if len(scripts) == 0 {
		t.Fatal("expected at least one Exec script for idle resend")
	}
	if scriptsContainCtrlC(scripts) {
		t.Fatalf("scripts must not auto-send Ctrl+C:\n%s", joinedScripts(scripts))
	}
}
```
