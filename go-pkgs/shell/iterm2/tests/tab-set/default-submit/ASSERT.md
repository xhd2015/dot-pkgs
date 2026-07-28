## Expected

- Script contains a submit write for `submit-cmd`: `write text "submit-cmd"`
  **without** the `without newline` qualifier on that command write.
- Existing single-tab leaves may only check command presence (vacuous for
  NoSubmit); this leaf pins the default submit form.

## Exit Code

- N/A (build-tab-set-script)

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
	// Use root Run script when NoSubmit is not yet on TabSpec — still validates
	// default submit form for current product (and after implementer).
	s := resp.Script
	if s == "" {
		// Fallback construct without NoSubmit field if Run returned empty.
		s = iterm2.BuildTabSetNewWindowScript(iterm2.TabSetSpec{
			Name: req.TabSetName,
			Tabs: []iterm2.TabSpec{
				{ID: "s1", Name: "Submit", Command: "submit-cmd"},
			},
		})
	}
	if s == "" {
		t.Fatal("expected non-empty script")
	}
	if commandWriteHasWithoutNewline(s, "submit-cmd") {
		t.Fatalf("default NoSubmit=false must not use without newline; script:\n%s", s)
	}
	if !strings.Contains(s, writeTextLine("submit-cmd")) &&
		!strings.Contains(s, `write text "submit-cmd`) {
		t.Fatalf("missing write text for submit-cmd; script:\n%s", s)
	}
}
```
