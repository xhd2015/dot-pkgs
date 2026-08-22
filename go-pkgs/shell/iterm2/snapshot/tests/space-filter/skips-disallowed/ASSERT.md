## Expected

- Only Desk2 deep-captured; FixedSpace pinned to 2.
- SpaceSkipped=1 for Desk0.

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
	if w.Index != 2 || w.Name != "Desk2" || w.WindowID != 20 {
		t.Fatalf("window %+v", w)
	}
	if w.FixedSpace == nil || *w.FixedSpace != 2 {
		t.Fatalf("FixedSpace=%v want 2", w.FixedSpace)
	}
}
```
