## Expected

- `err == nil`.
- `resp.AppPath == filepath.Join(req.Home, "Applications", "iTerm.app")`.
- App layout exists; post-condition `VerifyInstalled` succeeds.
- **`len(resp.OpenCalls) == 0`** — Open not called when flag is false.
- **`len(resp.ClearCalls) == 0`** — ClearQuarantineFn not called when flag is false.
- Register still runs (pipeline unchanged for non-open path).

## Errors

- None.

```go
import (
	"path/filepath"
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
	assertDirExists(t, want)
	assertFileExists(t, filepath.Join(want, "Contents", "Info.plist"))
	if len(resp.OpenCalls) != 0 {
		t.Fatalf("OpenCalls = %#v, want empty (InstallViaUserOpen=false)", resp.OpenCalls)
	}
	if len(resp.ClearCalls) != 0 {
		t.Fatalf("ClearCalls = %#v, want empty (InstallViaUserOpen=false)", resp.ClearCalls)
	}
	if resp.RegisterCalls < 1 {
		t.Fatalf("RegisterCalls = %d, want >= 1", resp.RegisterCalls)
	}
	if vErr := install.VerifyInstalled(resp.AppPath); vErr != nil {
		t.Fatalf("VerifyInstalled after InstallLatest: %v", vErr)
	}
}
```
