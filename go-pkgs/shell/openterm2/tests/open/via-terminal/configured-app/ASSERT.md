## Expected

- `err == nil`.
- `resp.Via == openterm2.ViaTerminal`.
- `resp.AppPath` equals the configured `TerminalApp`, not the default
  `/Applications/Utilities/Terminal.app`.
- `OpenTerminal` is called once with `req.Dir`.
- `OpenITerm` is not called.

## Side Effects

- Only the Terminal opener hook runs.

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
	assertEqual(t, "Via", resp.Via, openterm2.ViaTerminal)
	assertEqual(t, "AppPath", resp.AppPath, req.TerminalApp)
	if resp.AppPath == defaultTerminalApp {
		t.Fatalf("AppPath used default %q; want configured %q", defaultTerminalApp, req.TerminalApp)
	}
	assertOpenTerminalOnce(t, resp, req.Dir)
	assertNoOpenITerm(t, resp)
}
```
