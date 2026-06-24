## Expected

- `resp.ExitCode` is 0.
- `resp.Stderr` is empty.
- `resp.Stdout` trimmed lines are exactly:
  - `alice/alpha\towned`
  - `alice/beta\towned`
  in ascending sort order.

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
	want := "alice/alpha\towned\nalice/beta\towned\n"
	got := resp.Stdout
	if got != want {
		t.Fatalf("stdout mismatch:\nwant %q\ngot  %q", want, got)
	}
}
```