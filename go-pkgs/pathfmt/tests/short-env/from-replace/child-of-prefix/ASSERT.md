## Expected

- `resp.Display` equals `$X` + path remainder under the alias prefix.
- Remainder includes `pkg` and `a` with native separators.

## Errors

- `err` is nil.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	prefix := strings.TrimPrefix(req.Env[0], "X=")
	abs := absPath(t, req.Path)
	want := expectedEnvDisplay("X", prefix, abs)
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (path=%q)", want, resp.Display, req.Path)
	}
	if !strings.HasPrefix(resp.Display, "$X") {
		t.Fatalf("expected $X prefix, got %q", resp.Display)
	}
	if !strings.Contains(resp.Display, "pkg") || !strings.Contains(resp.Display, "a") {
		t.Fatalf("expected pkg/a remainder, got %q", resp.Display)
	}
}
```
