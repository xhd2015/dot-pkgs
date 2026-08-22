## Expected

- Snapshot has one window with App == `~/Applications/iTerm.app`.
- Idle enrich still succeeds.

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
	if snap.Windows[0].App != "~/Applications/iTerm.app" {
		t.Fatalf("App=%q", snap.Windows[0].App)
	}
}
```
