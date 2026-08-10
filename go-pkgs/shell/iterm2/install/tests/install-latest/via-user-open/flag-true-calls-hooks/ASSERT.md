## Expected

- `err == nil`.
- `resp.AppPath` ends with `iTerm.app` under Home Applications.
- **`len(resp.ClearCalls) == 1`** and `resp.ClearCalls[0] == resp.AppPath`.
- **`len(resp.OpenCalls) == 1`** and `resp.OpenCalls[0] == resp.AppPath`.
- Register called at least once with the same path.
- `VerifyInstalled(resp.AppPath)` succeeds.

## Errors

- None.

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
	assertNoError(t, err)
	want := filepath.Join(req.Home, "Applications", install.AppBundleName)
	assertEqual(t, "AppPath", resp.AppPath, want)
	if !strings.HasSuffix(resp.AppPath, install.AppBundleName) {
		t.Fatalf("AppPath %q does not end with %q", resp.AppPath, install.AppBundleName)
	}
	if len(resp.ClearCalls) != 1 {
		t.Fatalf("ClearCalls = %#v, want exactly 1 call", resp.ClearCalls)
	}
	if resp.ClearCalls[0] != want {
		t.Fatalf("ClearCalls[0] = %q, want %q", resp.ClearCalls[0], want)
	}
	if len(resp.OpenCalls) != 1 {
		t.Fatalf("OpenCalls = %#v, want exactly 1 call", resp.OpenCalls)
	}
	if resp.OpenCalls[0] != want {
		t.Fatalf("OpenCalls[0] = %q, want %q", resp.OpenCalls[0], want)
	}
	if resp.RegisterCalls < 1 {
		t.Fatalf("RegisterCalls = %d, want >= 1", resp.RegisterCalls)
	}
	if resp.RegisteredPath != want {
		t.Fatalf("RegisteredPath = %q, want %q", resp.RegisteredPath, want)
	}
	if vErr := install.VerifyInstalled(resp.AppPath); vErr != nil {
		t.Fatalf("VerifyInstalled after InstallLatest: %v", vErr)
	}
}
```
