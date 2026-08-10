## Expected

- Resolved path is `filepath.Join(HomeDir, "Applications", "iTerm.app")`.
- Must not prefer system when home exists.

## Exit Code

- N/A (library)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := homeApp(req.HomeDir)
	if resp.ResolvedPath != want {
		t.Fatalf("ResolveAppPathWith() = %q, want home %q", resp.ResolvedPath, want)
	}
}
```
