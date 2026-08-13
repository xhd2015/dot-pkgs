## Expected

- `err == nil`.
- `resp.Via == openterm2.ViaITerm2` (`"iterm2"`).
- `resp.AppPath` equals the injected resolve path.
- `OpenITerm` is called once with `req.Dir`.
- `OpenTerminal` is not called.

## Side Effects

- Only the iTerm opener hook runs.

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
	assertOpenITermOnce(t, resp, req.Dir)
	assertNoOpenTerminal(t, resp)
}
```
