## Expected

- Both `ScanStream` and `Scan` return no error.
- `Streamed` (walk order, lexical DFS) emits exactly `[., a/b, a-c]` — `a/b` before `a-c`,
  unsorted (DFS descends into `a/` before visiting sibling `a-c`).
- `Modules` (Scan, sorted by `Dir` lexical) is exactly `[., a-c, a/b]` — `a-c` before
  `a/b` because `-` (0x2D) < `/` (0x2F) at the 2nd character.
- The two orders differ, proving `ScanStream` does not sort.
- Both contain the same set of modules (root, `a/b`, `a-c`) with correct paths.

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
		t.Fatalf("ScanStream(%q) failed: %v", req.RootDir, resp.Err)
	}

	streamedDirs := dirLines(resp.Streamed)
	scanDirs := dirLines(resp.Modules)

	// Stream must be walk order (lexical DFS): a/b before a-c (no sort).
	wantStream := []string{".", "a/b", "a-c"}
	if !reflect.DeepEqual(streamedDirs, wantStream) {
		t.Fatalf("ScanStream dirs = %v, want %v (walk order, unsorted)", streamedDirs, wantStream)
	}

	// Scan must be sorted by Dir (lexical): a-c before a/b because '-' < '/'.
	wantScan := []string{".", "a-c", "a/b"}
	if !reflect.DeepEqual(scanDirs, wantScan) {
		t.Fatalf("Scan dirs = %v, want %v (sorted by Dir)", scanDirs, wantScan)
	}

	// Contrast: the two orders must differ, proving stream did not sort.
	if reflect.DeepEqual(streamedDirs, scanDirs) {
		t.Fatalf("stream order == scan order (%v); expected stream to be unsorted walk order", streamedDirs)
	}

	// Same set, correct paths.
	if p := pathOf(resp.Streamed, "a/b"); p != "example.com/root/a/b" {
		t.Fatalf("Path for 'a/b' = %q, want example.com/root/a/b", p)
	}
	if p := pathOf(resp.Streamed, "a-c"); p != "example.com/root/a-c" {
		t.Fatalf("Path for 'a-c' = %q, want example.com/root/a-c", p)
	}
}
```
