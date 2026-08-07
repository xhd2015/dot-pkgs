## Expected

- `err == nil`.
- Target app dir exists with `Contents/Info.plist` and MacOS binary.
- `VerifyInstalled`-compatible layout (plist + binary present).

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertDirExists(t, resp.AppPath)
	assertFileExists(t, filepath.Join(resp.AppPath, "Contents", "Info.plist"))
	assertFileExists(t, filepath.Join(resp.AppPath, "Contents", "MacOS", "iTerm2"))
}
```
