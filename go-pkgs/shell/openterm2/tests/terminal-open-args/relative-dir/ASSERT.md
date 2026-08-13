## Expected

- `err == nil`.
- `resp.TerminalArgs` equals
  `["open", "-a", "/Applications/Utilities/Terminal.app", filepath.Abs("rel-project")]`.
- The last element is absolute.

## Side Effects

- None (pure helper; no exec).

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertTerminalArgs(t, resp.TerminalArgs, defaultTerminalApp, req.Dir)
	if len(resp.TerminalArgs) != 4 || !filepath.IsAbs(resp.TerminalArgs[3]) {
		t.Fatalf("last argv element must be absolute; got %#v", resp.TerminalArgs)
	}
}
```
