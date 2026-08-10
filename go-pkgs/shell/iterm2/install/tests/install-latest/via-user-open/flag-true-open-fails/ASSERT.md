## Expected

- `err != nil` (Open failure aborts InstallLatest).
- **`len(resp.OpenCalls) == 1`** — Open was invoked once before failing.
- **`len(resp.ClearCalls) >= 1`** — clear runs before open (may still have been called).
- Recorded paths (when present) end with `iTerm.app` under Home Applications.

## Errors

- Non-nil error from InstallLatest (injected open failure).

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/install"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertError(t, err)
	want := filepath.Join(req.Home, "Applications", install.AppBundleName)
	if len(resp.OpenCalls) != 1 {
		t.Fatalf("OpenCalls = %#v, want exactly 1 call (then fail)", resp.OpenCalls)
	}
	if resp.OpenCalls[0] != want && !strings.HasSuffix(resp.OpenCalls[0], install.AppBundleName) {
		t.Fatalf("OpenCalls[0] = %q, want path ending with %q (prefer %q)",
			resp.OpenCalls[0], install.AppBundleName, want)
	}
	// Clear before open: ClearQuarantineFn may (and should) have run.
	if len(resp.ClearCalls) < 1 {
		t.Fatalf("ClearCalls = %#v, want >= 1 (clear before open)", resp.ClearCalls)
	}
	if resp.ClearCalls[0] != want && !strings.HasSuffix(resp.ClearCalls[0], install.AppBundleName) {
		t.Fatalf("ClearCalls[0] = %q, want path ending with %q (prefer %q)",
			resp.ClearCalls[0], install.AppBundleName, want)
	}
}
```
