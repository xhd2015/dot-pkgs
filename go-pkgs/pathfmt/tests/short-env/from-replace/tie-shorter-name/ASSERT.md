## Expected

- `resp.Display` uses `$X` (shorter name), not `$PROJECT_X`.

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
	var prefix string
	for _, e := range req.Env {
		if strings.HasPrefix(e, "X=") {
			prefix = strings.TrimPrefix(e, "X=")
			break
		}
	}
	abs := absPath(t, req.Path)
	want := expectedEnvDisplay("X", prefix, abs)
	if resp.Display != want {
		t.Fatalf("expected shorter name %q, got %q", want, resp.Display)
	}
	assertNoDollarVar(t, resp.Display, "PROJECT_X")
}
```
