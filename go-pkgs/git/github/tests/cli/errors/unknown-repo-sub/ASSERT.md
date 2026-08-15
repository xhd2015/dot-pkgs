## Expected

- `resp.ExitCode` is non-zero.
- `resp.Stderr` contains `unrecognized` (case-insensitive).
- `resp.Stdout` is empty.

## Side Effects

- No `gh` invocation.

## Errors

- Harness `err` is nil.

## Exit Code

- Non-zero

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	combined := strings.ToLower(resp.Stderr + resp.Stdout)
	if !strings.Contains(combined, "unrecognized") {
		t.Fatalf("expected unrecognized repo error, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```