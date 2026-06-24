## Expected

- `err` is nil.
- `len(resp.Results)` is 0 (empty slice, not nil error).
- Captured gh argv contains `--limit` and `30`.

## Side Effects

- Mock `gh repo list` invoked with default limit 30.

## Errors

- None.

## Exit Code

- N/A (library call).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Results == nil {
		t.Fatal("expected non-nil empty results slice")
	}
	if len(resp.Results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(resp.Results))
	}
	argv := resp.GhArgv
	if argv == "" {
		argv = readGhArgv(t, req.GhBin)
	}
	if !strings.Contains(argv, "--limit") {
		t.Fatalf("expected --limit in gh argv, got %q", argv)
	}
	if !strings.Contains(argv, "30") {
		t.Fatalf("expected default limit 30 in gh argv, got %q", argv)
	}
}
```