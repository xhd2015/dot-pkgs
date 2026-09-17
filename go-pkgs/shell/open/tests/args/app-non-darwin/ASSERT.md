## Expected

- `err != nil`.
- Error names `open -a` and the platform.

## Side Effects

- None: no argv is produced.

## Errors

- `open: open -a is unsupported on linux`.

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
	if !strings.Contains(err.Error(), "open -a") || !strings.Contains(err.Error(), "linux") {
		t.Fatalf("err = %v, want an open -a unsupported error naming linux", err)
	}
	if len(resp.Args) != 0 {
		t.Fatalf("Args = %#v, want none", resp.Args)
	}
}
```
