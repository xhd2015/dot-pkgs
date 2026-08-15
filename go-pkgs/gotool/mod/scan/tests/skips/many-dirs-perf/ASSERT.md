## Expected

- `Scan` returns no error.
- Exactly 1 module: the root (`Dir == "."`).
- `resp.Elapsed < 500ms` — the scan must complete within the performance budget.
  The `Run` function records wall-clock time in `resp.Elapsed` (via `time.Since(start)`
  wrapped around the `scan.Scan` call).

```go
import (
	"fmt"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Scan(%q) failed: %v", req.RootDir, resp.Err)
	}
	if len(resp.Modules) != 1 {
		t.Fatalf("Scan returned %d modules, want 1: %+v", len(resp.Modules), resp.Modules)
	}
	if resp.Modules[0].Dir != "." {
		t.Fatalf("only module Dir = %q, want \".\"", resp.Modules[0].Dir)
	}

	maxDuration := 500 * time.Millisecond
	if resp.Elapsed > maxDuration {
		t.Fatalf("Scan took %v over %d directories, want <%v (per-directory git check-ignore overhead is too high)",
			resp.Elapsed, 1+100, maxDuration)
	}
	t.Logf("Scan completed in %v (OK: <%v)", resp.Elapsed, maxDuration)

	// Ensure that Elapsed is actually set (the Run function was updated).
	if resp.Elapsed == 0 {
		t.Fatal("resp.Elapsed is 0 — Run function may not be recording timing")
	}
}

var _ = fmt.Sprintf
```
