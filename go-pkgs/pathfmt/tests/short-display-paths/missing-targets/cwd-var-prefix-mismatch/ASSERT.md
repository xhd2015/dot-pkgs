## Expected

- `resp.Display` is `.codex/hooks.json` (platform-native separators).
- Display must not contain `/var/folders/` or `/private/var/folders/` temp prefixes.

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
	want := filepath.Join(".codex", "hooks.json")
	if resp.Display != want {
		t.Fatalf("expected %q, got %q (cwd=%q path=%q)", want, resp.Display, resp.Cwd, req.Path)
	}
	for _, leak := range []string{"/var/folders/", "/private/var/folders/", "/tmp/"} {
		if strings.Contains(resp.Display, leak) {
			t.Fatalf("display leaked absolute temp prefix %q: %q", leak, resp.Display)
		}
	}
}
```