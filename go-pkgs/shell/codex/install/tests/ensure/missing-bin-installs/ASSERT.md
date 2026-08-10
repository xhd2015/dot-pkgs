## Expected

- `err == nil`.
- `resp.Action == "install"`.
- `resp.ShellCalls` is exactly `[install.InstallCmd]`.
- `resp.FetchLatestCalls == 0` (latest only when bin present).

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertEqual(t, "Action", resp.Action, "install")
	assertShellCalls(t, resp.ShellCalls, install.InstallCmd)
	assertEqual(t, "FetchLatestCalls", resp.FetchLatestCalls, 0)
}
```
