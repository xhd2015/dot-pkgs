## Expected

- Summary: Sessions=1, Unknown=1, Idle=0, Busy=0.
- Session `Idle` is nil (unknown).
- Warnings include a soft notice about no processes (substring `no processes`
  or the tty name).

## Errors

- `err` is nil (soft path; Capture still succeeds).

```go
import (
	"strings"
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
		Windows: 1, Tabs: 1, Sessions: 1, Idle: 0, Busy: 0, Unknown: 1,
	})
	s := snap.Windows[0].Tabs[0].Sessions[0]
	if s.Idle != nil {
		t.Fatalf("want Idle=nil (unknown), got %#v", s.Idle)
	}
	if len(resp.Warnings) == 0 {
		t.Fatal("expected soft warning for empty process list")
	}
	joined := strings.Join(resp.Warnings, "\n")
	if !strings.Contains(joined, "no processes") && !strings.Contains(joined, "ttys004") {
		t.Fatalf("warnings missing no-process soft path: %v", resp.Warnings)
	}
}
```
