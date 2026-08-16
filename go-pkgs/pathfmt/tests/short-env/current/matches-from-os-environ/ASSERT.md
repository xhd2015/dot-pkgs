## Expected

- `resp.Display` equals `pathfmt.ShortEnvFrom(req.Path, os.Environ())`.
- Do **not** assert a particular host `$VAR` value.

## Errors

- `err` is nil.

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := pathfmt.ShortEnvFrom(req.Path, os.Environ())
	if resp.Display != want {
		t.Fatalf("ShortEnv must equal ShortEnvFrom(..., os.Environ()): got %q want %q path=%q",
			resp.Display, want, req.Path)
	}
}
```
