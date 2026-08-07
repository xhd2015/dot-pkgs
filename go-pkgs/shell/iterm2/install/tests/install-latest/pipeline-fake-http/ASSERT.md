## Expected

- `err == nil`.
- `resp.Version == "3.6.11"`.
- `resp.AppPath == filepath.Join(req.Home, "Applications", "iTerm.app")`.
- App layout exists; Info.plist present.
- `resp.URL` has no arch substrings and mentions the zip name.
- Injected `Register` was called at least once with the install path.
- Post-condition: `install.VerifyInstalled(resp.AppPath)` succeeds.

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
	assertEqual(t, "Version", resp.Version, "3.6.11")
	want := filepath.Join(req.Home, "Applications", install.AppBundleName)
	assertEqual(t, "AppPath", resp.AppPath, want)
	assertDirExists(t, want)
	assertFileExists(t, filepath.Join(want, "Contents", "Info.plist"))
	assertFileExists(t, filepath.Join(want, "Contents", "MacOS", "iTerm2"))
	if !strings.Contains(resp.URL, "iTerm2-3_6_11.zip") {
		t.Fatalf("Result.URL %q missing zip name", resp.URL)
	}
	assertNoArchInURL(t, resp.URL)
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
