## Expected

- Tab `t1` Action is `resent`.
- Some Exec script contains `write text "resend-staged" without newline`
  (not only the bare command / bare write text without the qualifier).
- No Ctrl+C in Exec scripts.
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
	tab := iterm2.TabSpec{ID: "t1", Name: "Idle", Command: "resend-staged"}
	mustSetNoSubmit(t, &tab, true)
	spec := iterm2.TabSetSpec{
		Name: req.TabSetName,
		Tabs: []iterm2.TabSpec{tab},
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
	act := actionForTab(result, "t1")
	if act != "resent" {
		t.Fatalf("t1 Action = %q, want resent", act)
	}
	all := joinedScripts(scripts)
	want := writeTextWithoutNewline("resend-staged")
	if !strings.Contains(all, want) {
		escaped := iterm2.EscapeCommandForAppleScript("resend-staged")
		wantEsc := `write text "` + escaped + `" without newline`
		if !strings.Contains(all, wantEsc) {
			t.Fatalf("resend NoSubmit must emit without newline; scripts:\n%s", all)
		}
	}
	if scriptsContainCtrlC(scripts) {
		t.Fatalf("resend must not send Ctrl+C; scripts:\n%s", all)
	}
}
```
