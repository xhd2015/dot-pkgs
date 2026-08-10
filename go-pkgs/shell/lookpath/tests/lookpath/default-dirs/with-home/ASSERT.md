## Expected

- `err == nil`.
- `resp.DefaultDirs` contains, in order for the home-relative pair:
  1. `$home/.local/bin`
  2. `$home/go/bin`
  then system bins `/opt/homebrew/bin` and `/usr/local/bin` (any order relative
  to each other after home bins is acceptable if product documents order;
  **assert exact order**: home `.local/bin`, home `go/bin`, `/opt/homebrew/bin`,
  `/usr/local/bin`).

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
	assertNoError(t, err)
	home := req.DefaultDirsHome
	want := []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, "go", "bin"),
		"/opt/homebrew/bin",
		"/usr/local/bin",
	}
	if len(resp.DefaultDirs) < len(want) {
		t.Fatalf("DefaultDirs = %#v, want at least %#v", resp.DefaultDirs, want)
	}
	// Require the four contracted entries appear in this relative order.
	idx := 0
	for _, got := range resp.DefaultDirs {
		if idx < len(want) && filepath.Clean(got) == filepath.Clean(want[idx]) {
			idx++
		}
	}
	if idx != len(want) {
		t.Fatalf("DefaultDirs = %#v missing ordered contract %#v (matched %d)", resp.DefaultDirs, want, idx)
	}
}
```
