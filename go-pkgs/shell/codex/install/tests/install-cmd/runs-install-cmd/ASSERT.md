## Expected

- `err == nil`.
- `resp.ShellCalls` is exactly one entry equal to `install.InstallCmd`
  (`curl -fsSL https://chatgpt.com/codex/install.sh | sh`).

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
	assertShellCalls(t, resp.ShellCalls, install.InstallCmd)
}
```
