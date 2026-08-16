## Expected

- `resp.Display` uses `$AA` (alphabetically first), not `$BB`.

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
		if strings.HasPrefix(e, "AA=") {
			prefix = strings.TrimPrefix(e, "AA=")
			break
		}
	}
	abs := absPath(t, req.Path)
	want := expectedEnvDisplay("AA", prefix, abs)
	if resp.Display != want {
		t.Fatalf("expected alphabetical name %q, got %q", want, resp.Display)
	}
	assertNoDollarVar(t, resp.Display, "BB")
}
```
