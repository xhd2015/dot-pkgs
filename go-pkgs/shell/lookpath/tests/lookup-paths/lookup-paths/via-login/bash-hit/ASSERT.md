## Expected

- `err == nil`.
- One found item; Path = bash login path; `From == "bash"` (not `login_shell:bash`).
- RunLogin was called with shell `bash`.
- Item invariants hold.

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
	assertItemInvariants(t, resp.Items)
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}
	want := filepath.Join(req.WorkDir, "login", "bash", "mytool")
	assertItemFound(t, resp.Items[0], "mytool", want, "bash")
	if len(resp.RunLoginCalls) == 0 || resp.RunLoginCalls[0] != "bash" {
		t.Fatalf("RunLoginCalls = %#v, want first shell bash", resp.RunLoginCalls)
	}
}
```
