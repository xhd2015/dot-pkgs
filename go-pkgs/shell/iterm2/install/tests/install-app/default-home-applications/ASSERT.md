## Expected

- `err == nil`.
- `resp.AppPath == filepath.Join(req.Home, "Applications", "iTerm.app")`.
- App layout present under that path.

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
}
```
