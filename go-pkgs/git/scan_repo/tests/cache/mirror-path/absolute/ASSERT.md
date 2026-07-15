## Expected

- `resp.MirrorPath` equals
  `filepath.Join(CacheRoot, "mirror", "Users", "xhd2015", "Projects", "foo", "entry.json")`.
- Path ends with `entry.json` (not a hidden name).
- Path contains `/mirror/` (or OS separator equivalent) and does **not** contain
  an empty segment after `mirror` (no `mirror//Users`).

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
	want := filepath.Join(req.CacheRoot, "mirror", "Users", "xhd2015", "Projects", "foo", "entry.json")
	if resp.MirrorPath != want {
		t.Fatalf("MirrorPath = %q, want %q", resp.MirrorPath, want)
	}
	if filepath.Base(resp.MirrorPath) != "entry.json" {
		t.Fatalf("entry basename = %q, want entry.json", filepath.Base(resp.MirrorPath))
	}
	// No empty segment: "mirror" + "" + "Users" would produce mirror//Users or join quirks
	bad := string(filepath.Separator) + "mirror" + string(filepath.Separator) + string(filepath.Separator)
	if strings.Contains(resp.MirrorPath, bad) {
		t.Fatalf("MirrorPath has empty segment after mirror: %q", resp.MirrorPath)
	}
}
```
