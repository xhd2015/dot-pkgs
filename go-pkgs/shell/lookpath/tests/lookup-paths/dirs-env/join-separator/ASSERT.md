## Expected

- `err == nil`.
- Two found items with distinct parent dirs.
- `DirsEnv` equals join of those dirs with `os.PathListSeparator` in first-seen order.
- Item invariants hold.

## Errors

- None.

```go
import (
	"os"
	"path/filepath"
	"strings"
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
	dirA := filepath.Clean(filepath.Join(req.WorkDir, "bin-a"))
	dirB := filepath.Clean(filepath.Join(req.WorkDir, "bin-b"))
	want := strings.Join([]string{dirA, dirB}, string(os.PathListSeparator))
	// Normalize path segments in DirsEnv for comparison.
	parts := strings.Split(resp.DirsEnv, string(os.PathListSeparator))
	if len(parts) != 2 {
		t.Fatalf("DirsEnv = %q, want two parts joined by PathListSeparator", resp.DirsEnv)
	}
	for i := range parts {
		parts[i] = filepath.Clean(parts[i])
	}
	got := strings.Join(parts, string(os.PathListSeparator))
	assertEqual(t, "DirsEnv", got, want)
	if !strings.Contains(resp.DirsEnv, string(os.PathListSeparator)) {
		t.Fatalf("DirsEnv %q must contain PathListSeparator", resp.DirsEnv)
	}
}
```
