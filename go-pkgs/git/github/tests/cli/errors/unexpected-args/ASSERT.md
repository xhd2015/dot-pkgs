## Expected

- `resp.ExitCode` is non-zero.
- `resp.Stderr` contains `unexpected` (case-insensitive).
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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	msg := strings.ToLower(resp.Stderr + resp.Stdout)
	if !strings.Contains(msg, "unexpected") {
		t.Fatalf("expected unexpected arguments error, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```