## Expected

- `StatusTabSet` succeeds.
- States: `t-run`→`running`, `t-idle`→`idle`, `t-unk`→`unknown`, `t-miss`→`missing`.
- `SetName` is `bots`; `Instances` ≥ 1.

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
	tabs := make([]iterm2.TabSpec, len(req.Tabs))
	for i, tab := range req.Tabs {
		tabs[i] = iterm2.TabSpec{ID: tab.ID, Name: tab.Name, Command: tab.Command, Cwd: tab.Cwd}
	}
	spec := iterm2.TabSetSpec{Name: req.TabSetName, Tabs: tabs}

	st, serr := iterm2.StatusTabSet(spec, buildStatusConfig(req))
	if serr != nil {
		t.Fatalf("StatusTabSet: %v", serr)
	}
	if st == nil {
		t.Fatal("nil status")
	}
	if st.SetName != "bots" {
		t.Fatalf("SetName = %q, want bots", st.SetName)
	}
	want := map[string]string{
		"t-run":  "running",
		"t-idle": "idle",
		"t-unk":  "unknown",
		"t-miss": "missing",
	}
	for id, state := range want {
		got := statusStateFor(st, id)
		if got != state {
			t.Fatalf("tab %s State = %q, want %q; status=%+v", id, got, state, st)
		}
	}
	if st.Instances < 1 {
		t.Fatalf("Instances = %d, want >= 1", st.Instances)
	}
}
```
