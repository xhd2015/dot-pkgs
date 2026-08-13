## Expected

- `err == nil`.
- `resp.Via == openterm2.ViaTerminal` (`"terminal"`).
- `resp.AppPath` is `/Applications/Utilities/Terminal.app`.
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
	assertEqual(t, "AppPath", resp.AppPath, defaultTerminalApp)
	assertOpenTerminalOnce(t, resp, req.Dir)
	assertNoOpenITerm(t, resp)
}
```
