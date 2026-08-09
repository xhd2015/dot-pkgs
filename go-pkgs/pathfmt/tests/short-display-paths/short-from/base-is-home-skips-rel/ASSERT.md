## Expected

- `resp.Display` starts with `"~/"` and contains `".spl/seatalk-local-bot"`.
- `resp.Display` must **not** equal or start with `".spl/"` (cwd-style under home).

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
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.FromSlash("~/.spl/seatalk-local-bot/sessions/sid/SYSTEM.md")
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (base=home path=%q)", want, resp.Display, req.Path)
	}
	if strings.HasPrefix(resp.Display, ".spl") || strings.HasPrefix(resp.Display, ".spl"+string(filepath.Separator)) {
		t.Fatalf("base=home must skip rel; got cwd-style %q", resp.Display)
	}
}
```
