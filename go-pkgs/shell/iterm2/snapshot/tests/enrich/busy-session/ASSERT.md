## Expected

- Summary: Sessions=1, Busy=1, Idle=0, Unknown=0.
- Session `Idle` is false (busy).
- Chosen command reflects the busy leaf (`python` basename or full leaf in
  CommandLine); ShellPID set for the shell parent.
- Processes list is non-empty.

## Errors

- `err` is nil.

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
		Windows: 1, Tabs: 1, Sessions: 1, Idle: 0, Busy: 1, Unknown: 0,
	})
	s := snap.Windows[0].Tabs[0].Sessions[0]
	idle, ok := boolVal(s.Idle)
	if !ok || idle {
		t.Fatalf("want Idle=false, got %#v", s.Idle)
	}
	if s.ShellPID == nil {
		t.Fatal("expected ShellPID on busy session")
	}
	if s.PID == nil {
		t.Fatal("expected chosen PID on busy session")
	}
	if s.CommandLine == nil || !strings.Contains(*s.CommandLine, "python") {
		t.Fatalf("CommandLine want python leaf, got %#v", s.CommandLine)
	}
	if len(s.Processes) == 0 {
		t.Fatal("expected non-empty Processes")
	}
}
```
