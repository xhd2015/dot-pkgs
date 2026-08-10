## Expected

- `err == nil`.
- `resp.AppPath` basename is `iTerm.app`.
- `Contents/Info.plist` exists under app path.

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
	assertNoError(t, err)
	assertEqual(t, "AppBase", filepath.Base(resp.AppPath), install.AppBundleName)
	assertDirExists(t, resp.AppPath)
	assertFileExists(t, filepath.Join(resp.AppPath, "Contents", "Info.plist"))
}
```
