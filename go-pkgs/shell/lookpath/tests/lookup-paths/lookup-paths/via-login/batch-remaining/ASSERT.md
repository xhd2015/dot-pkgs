## Expected

- `err == nil`.
- Two found items in input order with Paths from login fixtures.
- Each item `From == "bash"`.
- RunLogin was invoked at least once (batch-friendly or per-name OK).
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
	if len(resp.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(resp.Items))
	}
	wantA := filepath.Join(req.WorkDir, "login", "bash", "tool-a")
	wantB := filepath.Join(req.WorkDir, "login", "bash", "tool-b")
	assertItemFound(t, resp.Items[0], "tool-a", wantA, "bash")
	assertItemFound(t, resp.Items[1], "tool-b", wantB, "bash")
	if len(resp.RunLoginCalls) == 0 {
		t.Fatal("expected RunLogin to be called for remaining names")
	}
	for _, sh := range resp.RunLoginCalls {
		if sh != "bash" && sh != "zsh" {
			t.Fatalf("unexpected RunLogin shell %q", sh)
		}
	}
}
```
