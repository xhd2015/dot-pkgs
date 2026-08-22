## Expected

- Summary: Windows=2, Tabs=3, Sessions=3, Idle=3.
- Window indices/names/IDs and tab indices/names preserved.
- Session IDs and TTYs match fixture; each has WindowIndex/TabIndex set.

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
		Windows: 2, Tabs: 3, Sessions: 3, Idle: 3, Busy: 0, Unknown: 0,
	})
	if len(snap.Windows) != 2 {
		t.Fatalf("windows=%d", len(snap.Windows))
	}
	w0, w1 := snap.Windows[0], snap.Windows[1]
	if w0.Index != 1 || w0.Name != "Win-A" || w0.WindowID != 10 {
		t.Fatalf("win0 %+v", w0)
	}
	if w1.Index != 2 || w1.Name != "Win-B" || w1.WindowID != 20 {
		t.Fatalf("win1 %+v", w1)
	}
	if len(w0.Tabs) != 1 || w0.Tabs[0].Index != 1 || w0.Tabs[0].Name != "A1" {
		t.Fatalf("win0 tabs %+v", w0.Tabs)
	}
	if len(w1.Tabs) != 2 {
		t.Fatalf("win1 tabs=%d", len(w1.Tabs))
	}
	if w1.Tabs[0].Index != 1 || w1.Tabs[0].Name != "B1" {
		t.Fatalf("tab B1 %+v", w1.Tabs[0])
	}
	if w1.Tabs[1].Index != 2 || w1.Tabs[1].Name != "B2" {
		t.Fatalf("tab B2 %+v", w1.Tabs[1])
	}
	sA := w0.Tabs[0].Sessions[0]
	sB1 := w1.Tabs[0].Sessions[0]
	sB2 := w1.Tabs[1].Sessions[0]
	if sA.ID != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" || sA.TTY != "/dev/ttys010" {
		t.Fatalf("sA %+v", sA)
	}
	if sB1.ID != "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" || sB1.TTY != "/dev/ttys020" {
		t.Fatalf("sB1 %+v", sB1)
	}
	if sB2.ID != "cccccccc-cccc-cccc-cccc-cccccccccccc" || sB2.TTY != "/dev/ttys021" {
		t.Fatalf("sB2 %+v", sB2)
	}
	if sA.WindowIndex != 1 || sA.TabIndex != 1 {
		t.Fatalf("sA layout w=%d t=%d", sA.WindowIndex, sA.TabIndex)
	}
	if sB2.WindowIndex != 2 || sB2.TabIndex != 2 {
		t.Fatalf("sB2 layout w=%d t=%d", sB2.WindowIndex, sB2.TabIndex)
	}
}
```
