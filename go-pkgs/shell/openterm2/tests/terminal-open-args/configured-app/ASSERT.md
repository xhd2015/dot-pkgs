## Expected

- `err == nil`.
- `resp.TerminalArgs` equals
  `["open", "-a", <custom app>, abs(ValidDir)]`.
- The app element is the override, not `/Applications/Utilities/Terminal.app`.

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
	assertTerminalArgs(t, resp.TerminalArgs, req.ArgsAppPath, req.Dir)
	if len(resp.TerminalArgs) >= 3 && resp.TerminalArgs[2] == defaultTerminalApp {
		t.Fatalf("argv app is default %q; want override %q", defaultTerminalApp, req.ArgsAppPath)
	}
}
```
