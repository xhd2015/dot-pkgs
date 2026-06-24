## Expected

- `resp.ExitCode` is 0.
- `resp.Stderr` is empty.
- `resp.Stdout` trimmed is `[]`.

## Side Effects

- Mock `gh api user` and `gh repo list alice` invoked.

## Errors

- Harness `err` is nil.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) != "[]" {
		t.Fatalf("expected empty JSON array [], got %q", resp.Stdout)
	}
}
```