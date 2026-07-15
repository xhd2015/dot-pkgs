## Expected

- `resp.MirrorPath` equals `expectedMirrorEntryPath(CacheRoot, RealPath)`.
- The same mapping for `filepath.Abs(RealPath)` yields the identical string
  (relative and absolute forms share one mirror key).
- Result ends with `entry.json` under `.../mirror/...`.

## Errors

- `err` is nil.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	want := expectedMirrorEntryPath(t, req.CacheRoot, req.RealPath)
	if resp.MirrorPath != want {
		t.Fatalf("MirrorPath = %q, want %q", resp.MirrorPath, want)
	}
	abs := absPath(t, req.RealPath)
	absMapped, mapErr := scan_repo.MirrorEntryPath(req.CacheRoot, abs)
	if mapErr != nil {
		t.Fatalf("MirrorEntryPath(abs): %v", mapErr)
	}
	if absMapped != resp.MirrorPath {
		t.Fatalf("relative map %q != abs map %q (abs real %q)", resp.MirrorPath, absMapped, abs)
	}
	if filepath.Base(resp.MirrorPath) != "entry.json" {
		t.Fatalf("basename = %q, want entry.json", filepath.Base(resp.MirrorPath))
	}
	mirrorRoot := filepath.Join(req.CacheRoot, "mirror")
	rel, relErr := filepath.Rel(mirrorRoot, resp.MirrorPath)
	if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Fatalf("MirrorPath %q not under %q", resp.MirrorPath, mirrorRoot)
	}
}
```

