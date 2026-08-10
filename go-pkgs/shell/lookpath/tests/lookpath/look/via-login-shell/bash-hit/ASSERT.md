## Expected

- `err == nil`.
- `resp.Path` equals the bash login path (newline trimmed).
- `resp.Via == "login_shell:bash"`.
- `resp.RunLoginCalls` starts with `"bash"`.
- Command string includes `command -v` and the binary name.

## Errors

- None.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	want := filepath.Join(req.WorkDir, "login", "bash", "mytool")
	assertPathEqual(t, resp.Path, want)
	assertEqual(t, "Via", resp.Via, "login_shell:bash")
	if len(resp.RunLoginCalls) == 0 || resp.RunLoginCalls[0] != "bash" {
		t.Fatalf("RunLoginCalls = %#v, want first shell bash", resp.RunLoginCalls)
	}
	if len(resp.RunLoginCommands) == 0 {
		t.Fatal("expected RunLogin command recorded")
	}
	cmd := resp.RunLoginCommands[0]
	if !strings.Contains(cmd, "command -v") || !strings.Contains(cmd, req.Name) {
		t.Fatalf("RunLogin command %q should contain command -v and name %q", cmd, req.Name)
	}
}
```
