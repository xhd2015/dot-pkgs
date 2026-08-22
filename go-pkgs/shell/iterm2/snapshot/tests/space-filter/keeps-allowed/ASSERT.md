## Expected

- One deep-captured window (Index=1, Name=Desk0); FixedSpace pinned to 0.
- Summary Windows=1, Sessions=1, Idle=1.
- SpaceSkipped=1 (the space-2 header).

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
	if resp.SpaceSkipped != 1 {
		t.Fatalf("SpaceSkipped=%d want 1", resp.SpaceSkipped)
	}
	if len(snap.Windows) != 1 {
		t.Fatalf("windows=%d", len(snap.Windows))
	}
	w := snap.Windows[0]
	if w.Index != 1 || w.Name != "Desk0" || w.WindowID != 10 {
		t.Fatalf("window %+v", w)
	}
	if w.FixedSpace == nil || *w.FixedSpace != 0 {
		t.Fatalf("FixedSpace=%v want 0", w.FixedSpace)
	}
	if len(w.Tabs) != 1 || len(w.Tabs[0].Sessions) != 1 {
		t.Fatalf("tabs/sessions %+v", w.Tabs)
	}
}
```
