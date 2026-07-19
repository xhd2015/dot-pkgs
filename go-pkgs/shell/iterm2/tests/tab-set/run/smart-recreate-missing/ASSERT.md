## Expected

- Tab `t2` Action is `created` (missing tab recreated in chosen window).
- Exec scripts include `create tab` and command `missing-cmd`.
- Markers for set/tab may appear for the new tab.

## Exit Code

- N/A

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
	act := actionForTab(result, "t2")
	if act != "created" && act != "missing-created" {
		t.Fatalf("t2 Action = %q, want created (or missing-created)", act)
	}
	all := joinedScripts(scripts)
	if !strings.Contains(all, `create tab`) {
		t.Fatalf("missing create tab for recreated tab; scripts:\n%s", all)
	}
	if !strings.Contains(all, "missing-cmd") {
		t.Fatalf("missing command for recreated tab; scripts:\n%s", all)
	}
}
```
