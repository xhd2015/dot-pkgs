## Expected

- `err == nil`.
- Two found items under the same parent dir.
- `len(Dirs) == 1` and equals that parent (cleaned).
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
	wantDir := filepath.Clean(req.ExtraDirs[0])
	if len(resp.Dirs) != 1 {
		t.Fatalf("Dirs = %#v, want single entry %q", resp.Dirs, wantDir)
	}
	assertEqual(t, "Dirs[0]", filepath.Clean(resp.Dirs[0]), wantDir)
}
```
