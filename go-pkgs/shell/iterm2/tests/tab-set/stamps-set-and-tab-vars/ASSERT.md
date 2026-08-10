## Expected

- Script mentions both session variable names:
  - `user.koolTabSet` (set id)
  - `user.koolTabSetTab` (per-tab id)
- Set name `bots` is assigned to `user.koolTabSet` (at least once per tab → ≥2).
- Tab IDs `t1` and `t2` appear as values for `user.koolTabSetTab`.
- Each of the two variable names appears at least twice (once per tab).

## Exit Code

- N/A (build-tab-set-script phase)

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {

	if err != nil {
		t.Fatal(err)
	}
	s := resp.Script
	if s == "" {
		t.Fatal("expected non-empty script")
	}
	setVar := `variable named "` + tabSetVarLiteral + `"`
	tabVar := `variable named "` + tabSetTabVarLiteral + `"`
	if !strings.Contains(s, setVar) {
		t.Fatalf("missing %s; script:\n%s", setVar, s)
	}
	if !strings.Contains(s, tabVar) {
		t.Fatalf("missing %s; script:\n%s", tabVar, s)
	}
	if n := countOccurrences(s, setVar); n < 2 {
		t.Fatalf("%s count = %d, want ≥2 (one per tab); script:\n%s", setVar, n, s)
	}
	if n := countOccurrences(s, tabVar); n < 2 {
		t.Fatalf("%s count = %d, want ≥2 (one per tab); script:\n%s", tabVar, n, s)
	}
	// set value "bots" near the set variable (assignment form flexible)
	if !strings.Contains(s, `"bots"`) {
		t.Fatalf("missing set name value bots; script:\n%s", s)
	}
	if !strings.Contains(s, `"t1"`) {
		t.Fatalf("missing tab id t1; script:\n%s", s)
	}
	if !strings.Contains(s, `"t2"`) {
		t.Fatalf("missing tab id t2; script:\n%s", s)
	}
}
```
