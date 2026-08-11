## Expected

- `err != nil` (Open failure aborts InstallLatest).
- **`len(resp.OpenCalls) == 1`** — Open was invoked once before failing.
- **`len(resp.ClearCalls) == 0`** — clear is not part of user-open.
- **`resp.RegisterCalls == 0`**.
- `resp.AppPath` is the staged extract path (ends with `iTerm.app`), not Home Applications.

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
	if len(resp.OpenCalls) != 1 {
		t.Fatalf("OpenCalls = %#v, want exactly 1 call (then fail)", resp.OpenCalls)
	}
	if !strings.HasSuffix(resp.OpenCalls[0], install.AppBundleName) {
		t.Fatalf("OpenCalls[0] = %q, want path ending with %q",
			resp.OpenCalls[0], install.AppBundleName)
	}
	placed := filepath.Join(req.Home, "Applications", install.AppBundleName)
	if resp.OpenCalls[0] == placed {
		t.Fatalf("OpenCalls[0] is place path %q; want staged extract", placed)
	}
	if len(resp.ClearCalls) != 0 {
		t.Fatalf("ClearCalls = %#v, want empty", resp.ClearCalls)
	}
	if resp.RegisterCalls != 0 {
		t.Fatalf("RegisterCalls = %d, want 0", resp.RegisterCalls)
	}
	if resp.AppPath != "" && resp.AppPath != resp.OpenCalls[0] {
		t.Fatalf("AppPath = %q, OpenCalls[0] = %q", resp.AppPath, resp.OpenCalls[0])
	}
}
```
