## Expected

- `err == nil`.
- One found item; Path = zsh login path; `From == "zsh"`.
- RunLogin call order includes bash then zsh.
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
	want := filepath.Join(req.WorkDir, "login", "zsh", "mytool")
	assertItemFound(t, resp.Items[0], "mytool", want, "zsh")
	if len(resp.RunLoginCalls) < 2 {
		t.Fatalf("RunLoginCalls = %#v, want bash then zsh", resp.RunLoginCalls)
	}
	assertEqual(t, "RunLoginCalls[0]", resp.RunLoginCalls[0], "bash")
	assertEqual(t, "RunLoginCalls[1]", resp.RunLoginCalls[1], "zsh")
}
```
