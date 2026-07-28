## Expected

- Script embeds cwd path `/tmp/stage-cwd` and a `cd` write for that path.
- The `cd` write must **not** use `without newline` (cd always executes).
- Command `staged-cwd-cmd` uses `write text "…" without newline`.
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
	_ = d

	if err != nil {
		t.Fatal(err)
	}
	tab := iterm2.TabSpec{
		ID:      "c1",
		Name:    "CwdStage",
		Command: "staged-cwd-cmd",
		Cwd:     "/tmp/stage-cwd",
	}
	mustSetNoSubmit(t, &tab, true)
	spec := iterm2.TabSetSpec{
		Name: req.TabSetName,
		Tabs: []iterm2.TabSpec{tab},
	}
	s := iterm2.BuildTabSetNewWindowScript(spec)
	if s == "" {
		t.Fatal("expected non-empty script")
	}
	const cwd = `/tmp/stage-cwd`
	if !strings.Contains(s, cwd) {
		t.Fatalf("missing cwd path %q; script:\n%s", cwd, s)
	}
	hasCdForm := strings.Contains(s, `write text ("cd "`) ||
		strings.Contains(s, `write text "cd `) ||
		(strings.Contains(s, "cd ") && strings.Contains(s, cwd))
	if !hasCdForm {
		t.Fatalf("expected cd write for Cwd; script:\n%s", s)
	}
	// cd must execute: no "without newline" on the cd line.
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, "write text") && strings.Contains(line, "cd ") &&
			strings.Contains(line, "without newline") {
			t.Fatalf("cd must execute with newline; line=%q\nscript:\n%s", line, s)
		}
	}
	if !commandWriteHasWithoutNewline(s, "staged-cwd-cmd") {
		escaped := iterm2.EscapeCommandForAppleScript("staged-cwd-cmd")
		wantEsc := `write text "` + escaped + `" without newline`
		if !strings.Contains(s, wantEsc) {
			t.Fatalf("command must use without newline; script:\n%s", s)
		}
	}
}
```
