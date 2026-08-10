## Expected

- Tab `t2` Action is `created` (or `missing-created`).
- Exec scripts include `create tab` and `write text "missing-staged" without newline`.
- NoSubmit set via reflection (compile-safe Classic TDD).

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
	t1 := iterm2.TabSpec{ID: "t1", Name: "One", Command: "cmd-one"}
	t2 := iterm2.TabSpec{ID: "t2", Name: "Two", Command: "missing-staged"}
	mustSetNoSubmit(t, &t2, true)
	spec := iterm2.TabSetSpec{
		Name: req.TabSetName,
		Tabs: []iterm2.TabSpec{t1, t2},
	}
	result, rerr := iterm2.RunTabSet(spec, iterm2.TabSetRunOptions{
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
	want := writeTextWithoutNewline("missing-staged")
	if !strings.Contains(all, want) {
		escaped := iterm2.EscapeCommandForAppleScript("missing-staged")
		wantEsc := `write text "` + escaped + `" without newline`
		if !strings.Contains(all, wantEsc) {
			t.Fatalf("create-tab NoSubmit must emit without newline; scripts:\n%s", all)
		}
	}
}
```