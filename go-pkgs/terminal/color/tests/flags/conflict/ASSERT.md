## Expected

- `err` is non-nil.
- `err.Error()` equals exactly `--color and --no-color cannot be specified together`.

## Side Effects

- None.

## Errors

- Exact string: `--color and --no-color cannot be specified together`.

## Exit Code

- N/A (in-process library).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("ModeFromFlags(true, true): want error, got nil")
	}
	const want = "--color and --no-color cannot be specified together"
	if got := err.Error(); got != want {
		t.Fatalf("ModeFromFlags(true, true): error %q, want %q", got, want)
	}
}
```
