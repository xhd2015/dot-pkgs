## Expected

- `resp.ExitCode` is 0.
- `resp.Stdout` is non-empty and mentions `repo` (available subcommand).
- `resp.Stderr` is empty.

## Side Effects

- No `gh` invocation (`req.GhBin` unset).

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
	if strings.TrimSpace(resp.Stdout) == "" {
		t.Fatal("expected non-empty stdout usage")
	}
	if !strings.Contains(strings.ToLower(resp.Stdout), "repo") {
		t.Fatalf("expected repo in usage, got stdout=%q", resp.Stdout)
	}
	if strings.TrimSpace(resp.Stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", resp.Stderr)
	}
}
```