## Expected

- `resp.Display` is `"~/.spl/seatalk-local-bot/sessions/sid/SYSTEM.md"` (native seps).
- Does not start with `".spl"`.

## Errors

- `err` is nil.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.FromSlash("~/.spl/seatalk-local-bot/sessions/sid/SYSTEM.md")
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (base=%q path=%q)", want, resp.Display, req.BaseDir, req.Path)
	}
	if strings.HasPrefix(resp.Display, ".spl") {
		t.Fatalf("must not be serve-cwd relative .spl form: %q", resp.Display)
	}
}
```
