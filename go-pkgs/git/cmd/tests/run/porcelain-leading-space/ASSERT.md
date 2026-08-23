## Expected

- `err` is nil.
- `resp.Output` starts with a space then `M` (` M README`), not `M ` (staged).

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
	out := resp.Output
	if !strings.HasPrefix(out, " M ") && !strings.HasPrefix(out, " M\t") {
		// Porcelain is " M path" — leading space is the index column.
		if len(out) < 2 || out[0] != ' ' || out[1] != 'M' {
			t.Fatalf("want leading-space unstaged M line, got %q", out)
		}
	}
}
```
