## Expected

- `Scan` returns no error.
- Exactly 2 modules: root (`Dir == "."`) and `untracked` (`Dir == "untracked"`).
- `untracked` is present (no git at root → git-based skips disabled; the dir is not
  `.git`/`vendor`/`testdata`, so it is scanned).
- Sorted by `Dir`: `.` then `untracked`.

```go
import (
	"reflect"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Scan(%q) failed: %v", req.RootDir, resp.Err)
	}

	gotDirs := dirLines(resp.Modules)
	wantDirs := []string{".", "untracked"}
	if !reflect.DeepEqual(gotDirs, wantDirs) {
		t.Fatalf("Scan dirs = %v, want %v (untracked must be included; no git at root)", gotDirs, wantDirs)
	}

	if p := pathOf(resp.Modules, "untracked"); p != "example.com/root/untracked" {
		t.Fatalf("Path for 'untracked' = %q, want example.com/root/untracked", p)
	}
}
```
