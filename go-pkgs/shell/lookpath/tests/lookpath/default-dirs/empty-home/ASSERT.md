## Expected

- `err == nil`.
- `resp.DefaultDirs` contains `/opt/homebrew/bin` and `/usr/local/bin`.
- Empty home must **not** invent home-relative joins such as `/.local/bin` or
  `/go/bin`.

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
	assertContainsAll(t, resp.DefaultDirs, "/opt/homebrew/bin", "/usr/local/bin")
	for _, bad := range []string{"/.local/bin", "/go/bin"} {
		for _, g := range resp.DefaultDirs {
			if filepath.Clean(g) == filepath.Clean(bad) {
				t.Fatalf("DefaultDirs contains empty-home join artifact %q; got %#v", g, resp.DefaultDirs)
			}
		}
	}
}
```
