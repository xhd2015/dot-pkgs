## Expected

- `err != nil`.
- Error mentions the platform, so an operator can see which one was refused.

## Side Effects

- None: no argv is produced.

## Errors

- `open: unsupported platform plan9`.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertError(t, err)
	if !strings.Contains(err.Error(), "unsupported platform") || !strings.Contains(err.Error(), "plan9") {
		t.Fatalf("err = %v, want an unsupported-platform error naming plan9", err)
	}
	if len(resp.Args) != 0 {
		t.Fatalf("Args = %#v, want none", resp.Args)
	}
}
```
