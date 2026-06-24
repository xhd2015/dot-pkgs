## Expected

- `resp.ExitCode` is 0.
- `resp.Stdout` mentions `list` and documents `--owner` and `--json`.
- `resp.Stderr` is empty.

## Side Effects

- No `gh` invocation.

## Errors

- `err` from harness is nil.

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
	out := strings.ToLower(resp.Stdout)
	for _, want := range []string{"list", "--owner", "--json"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in usage, got stdout=%q", want, resp.Stdout)
		}
	}
	if strings.TrimSpace(resp.Stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", resp.Stderr)
	}
}
```