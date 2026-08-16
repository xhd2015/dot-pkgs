## Expected

- `resp.Display` is `$AI` + remainder under the AI prefix (e.g. `$AI/src/foo`).
- Display does **not** use `$X`.

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
	var aiPrefix string
	for _, e := range req.Env {
		if strings.HasPrefix(e, "AI=") {
			aiPrefix = strings.TrimPrefix(e, "AI=")
		}
	}
	abs := absPath(t, req.Path)
	want := expectedEnvDisplay("AI", aiPrefix, abs)
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (path=%q env=%v)", want, resp.Display, req.Path, req.Env)
	}
	assertNoDollarVar(t, resp.Display, "X")
}
```
