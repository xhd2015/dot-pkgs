## Expected

- `err == nil`.
- `Via=iterm2` and `AppPath` is the iTerm resolve path, not `TerminalApp`.
- `OpenITerm` called once with `req.Dir`.
- `OpenTerminal` not called.

## Side Effects

- Only the iTerm opener hook runs; configured Terminal app is unused.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/openterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertEqual(t, "Via", resp.Via, openterm2.ViaITerm2)
	assertEqual(t, "AppPath", resp.AppPath, req.ITermAppPath)
	if resp.AppPath == req.TerminalApp {
		t.Fatalf("AppPath used TerminalApp %q; want iTerm path %q", req.TerminalApp, req.ITermAppPath)
	}
	assertOpenITermOnce(t, resp, req.Dir)
	assertNoOpenTerminal(t, resp)
}
```
