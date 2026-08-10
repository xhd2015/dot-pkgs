## Expected

- `err == nil`.
- `resp.Path` equals the zsh login path.
- `resp.Via == "login_shell:zsh"`.
- RunLogin call order includes bash then zsh.

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
	want := filepath.Join(req.WorkDir, "login", "zsh", "mytool")
	assertPathEqual(t, resp.Path, want)
	assertEqual(t, "Via", resp.Via, "login_shell:zsh")
	if len(resp.RunLoginCalls) < 2 {
		t.Fatalf("RunLoginCalls = %#v, want bash then zsh", resp.RunLoginCalls)
	}
	assertEqual(t, "RunLoginCalls[0]", resp.RunLoginCalls[0], "bash")
	assertEqual(t, "RunLoginCalls[1]", resp.RunLoginCalls[1], "zsh")
}
```
