## Expected

- `resp.MirrorPath` equals
  `filepath.Join(CacheRoot, "mirror", "Users", "xhd2015", "Projects", "org", "team", "repo", "entry.json")`.
- Intermediate path under `mirror/` preserves each real-path segment in order.

## Errors

- `err` is nil.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	want := filepath.Join(req.CacheRoot, "mirror", "Users", "xhd2015", "Projects", "org", "team", "repo", "entry.json")
	if resp.MirrorPath != want {
		t.Fatalf("MirrorPath = %q, want %q", resp.MirrorPath, want)
	}
	rel, err := filepath.Rel(filepath.Join(req.CacheRoot, "mirror"), filepath.Dir(resp.MirrorPath))
	if err != nil {
		t.Fatal(err)
	}
	// Expect Users/xhd2015/Projects/org/team/repo (OS separators)
	parts := strings.Split(rel, string(filepath.Separator))
	wantParts := []string{"Users", "xhd2015", "Projects", "org", "team", "repo"}
	if len(parts) != len(wantParts) {
		t.Fatalf("mirror relative segments = %v, want %v", parts, wantParts)
	}
	for i := range wantParts {
		if parts[i] != wantParts[i] {
			t.Fatalf("segment[%d] = %q, want %q (full %v)", i, parts[i], wantParts[i], parts)
		}
	}
}
```
