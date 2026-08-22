## Expected

- Summary: Windows=1, Tabs=1, Sessions=1, Idle=1, Busy=0, Unknown=0.
- Session preserves ID/TTY; `Idle` is true; WindowIndex/TabIndex set.
- No Agent field on session type (compile-time model contract).

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	snap := mustSnap(t, resp, err)
	assertSummary(t, snap.Summary, snapshot.SnapshotSummary{
		Windows: 1, Tabs: 1, Sessions: 1, Idle: 1, Busy: 0, Unknown: 0,
	})
	if len(snap.Windows) != 1 {
		t.Fatalf("windows=%d", len(snap.Windows))
	}
	w := snap.Windows[0]
	if w.Index != 1 || w.Name != "Main" || w.WindowID != 42 {
		t.Fatalf("window %+v", w)
	}
	if len(w.Tabs) != 1 || len(w.Tabs[0].Sessions) != 1 {
		t.Fatalf("tabs/sessions %+v", w.Tabs)
	}
	s := w.Tabs[0].Sessions[0]
	if s.ID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("ID=%q", s.ID)
	}
	if s.TTY != "/dev/ttys001" {
		t.Fatalf("TTY=%q", s.TTY)
	}
	idle, ok := boolVal(s.Idle)
	if !ok || !idle {
		t.Fatalf("want Idle=true, got %#v", s.Idle)
	}
	if s.WindowIndex != 1 || s.TabIndex != 1 {
		t.Fatalf("layout hints window=%d tab=%d", s.WindowIndex, s.TabIndex)
	}
	if s.ShellPID == nil {
		t.Fatal("expected ShellPID for idle shell session")
	}
}
```
