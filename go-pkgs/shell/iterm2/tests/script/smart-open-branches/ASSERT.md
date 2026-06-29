## Expected

- Script scans session `path`, includes tab reuse and window fallback.
- Script cds via `write text ("cd " & quoted form of targetDir)`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	s := resp.Script
	if !scriptHasPathScan(s) {
		t.Fatalf("missing path scan: %q", s)
	}
	if !strings.Contains(s, "create tab with default profile") {
		t.Fatalf("missing create tab: %q", s)
	}
	if !strings.Contains(s, "create window with default profile") {
		t.Fatalf("missing create window fallback: %q", s)
	}
	if !strings.Contains(s, `write text ("cd " & quoted form of targetDir)`) {
		t.Fatalf("missing cd write text: %q", s)
	}
}
```