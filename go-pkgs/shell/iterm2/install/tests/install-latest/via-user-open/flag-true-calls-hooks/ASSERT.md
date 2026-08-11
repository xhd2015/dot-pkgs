## Expected

- `err == nil`.
- `resp.AppPath` ends with `iTerm.app` under the **extract/cache** tree (staged),
  **not** under Home Applications.
- **`len(resp.OpenCalls) == 1`** and `resp.OpenCalls[0] == resp.AppPath`.
- **`len(resp.ClearCalls) == 0`** — user-open does not clear quarantine.
- **`resp.RegisterCalls == 0`** — user-open does not place/register.
- `VerifyInstalled(resp.AppPath)` succeeds on the staged bundle.
- Home Applications path must **not** exist as a placed install.

## Errors

- None.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/install"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	if resp.AppPath == "" {
		t.Fatal("AppPath empty, want staged extract path")
	}
	if !strings.HasSuffix(resp.AppPath, install.AppBundleName) {
		t.Fatalf("AppPath %q does not end with %q", resp.AppPath, install.AppBundleName)
	}
	// Must be staged (under cache/extract), not placed into Applications.
	placed := filepath.Join(req.Home, "Applications", install.AppBundleName)
	if resp.AppPath == placed {
		t.Fatalf("AppPath = %q is Home Applications place path; want staged extract", resp.AppPath)
	}
	if _, stErr := os.Stat(placed); stErr == nil {
		t.Fatalf("unexpected place at %s; user-open must not InstallApp", placed)
	}
	if len(resp.OpenCalls) != 1 {
		t.Fatalf("OpenCalls = %#v, want exactly 1 call", resp.OpenCalls)
	}
	if resp.OpenCalls[0] != resp.AppPath {
		t.Fatalf("OpenCalls[0] = %q, want AppPath %q", resp.OpenCalls[0], resp.AppPath)
	}
	if len(resp.ClearCalls) != 0 {
		t.Fatalf("ClearCalls = %#v, want empty (user-open does not clear quarantine)", resp.ClearCalls)
	}
	if resp.RegisterCalls != 0 {
		t.Fatalf("RegisterCalls = %d, want 0 (user-open does not register)", resp.RegisterCalls)
	}
	if vErr := install.VerifyInstalled(resp.AppPath); vErr != nil {
		t.Fatalf("VerifyInstalled after InstallLatest: %v", vErr)
	}
}
```
