## Expected

- Capture succeeds with non-nil snapshot.
- `Source` is `iterm2`; `Host` is the injected hostname (`testhost`).
- `Windows` is empty; summary Windows/Tabs/Sessions/Idle/Busy/Unknown all 0.
- `CapturedAt` is non-empty (from fixed Now).

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
	if snap.Source != "iterm2" {
		t.Fatalf("Source=%q want iterm2", snap.Source)
	}
	if snap.Host != "testhost" {
		t.Fatalf("Host=%q want testhost", snap.Host)
	}
	if snap.CapturedAt == "" {
		t.Fatal("CapturedAt empty")
	}
	if len(snap.Windows) != 0 {
		t.Fatalf("Windows len=%d want 0", len(snap.Windows))
	}
	assertSummary(t, snap.Summary, snapshot.SnapshotSummary{})
}
```
