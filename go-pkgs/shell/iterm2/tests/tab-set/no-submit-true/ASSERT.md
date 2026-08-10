## Expected

- AppleScript includes `write text "staged-cmd" without newline` (or escaped
  equivalent with trailing `without newline`).
- Must not only contain the command string without the without-newline qualifier
  as the sole write form for that command.
- TabSpec.NoSubmit is set via reflection so the leaf compiles before the field
  exists (missing field → assert RED via setNoSubmit).

## Exit Code

- N/A (build-tab-set-script; product invoked in Assert for NoSubmit wiring)

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
	// Compile-safe: no NoSubmit field in struct literal.
	tab := iterm2.TabSpec{ID: "s1", Name: "Stage", Command: "staged-cmd"}
	mustSetNoSubmit(t, &tab, true)
	spec := iterm2.TabSetSpec{
		Name: req.TabSetName,
		Tabs: []iterm2.TabSpec{tab},
	}
	s := iterm2.BuildTabSetNewWindowScript(spec)
	if s == "" {
		t.Fatal("expected non-empty script")
	}
	want := writeTextWithoutNewline("staged-cmd")
	if !strings.Contains(s, want) {
		// Accept EscapeCommandForAppleScript form if product escapes first.
		escaped := iterm2.EscapeCommandForAppleScript("staged-cmd")
		wantEsc := `write text "` + escaped + `" without newline`
		if !strings.Contains(s, wantEsc) {
			t.Fatalf("NoSubmit must emit %q (or %q); script:\n%s", want, wantEsc, s)
		}
	}
	// Presence of bare write text "staged-cmd" alone is insufficient when it is
	// only a prefix of the without-newline form — require the qualifier.
	if !commandWriteHasWithoutNewline(s, "staged-cmd") &&
		!strings.Contains(s, `without newline`) {
		t.Fatalf("missing without newline for staged-cmd; script:\n%s", s)
	}
}
```
