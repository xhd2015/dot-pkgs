## Expected

- `err == nil`.
- `resp.TerminalArgs` equals
  `["open", "-a", "/Applications/Utilities/Terminal.app", abs(ValidDir)]`.

## Side Effects

- None (pure helper; no exec).

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertTerminalArgs(t, resp.TerminalArgs, defaultTerminalApp, req.Dir)
}
```
