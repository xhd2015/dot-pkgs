## Expected

- `resp.ExitCode` is non-zero.
- `resp.Stderr` contains `gh auth login` (case-insensitive).
- `resp.Stdout` is empty.
- Mock `gh` invoked for `api user` only; `repo list` never runs.

## Side Effects

- Auth check runs before any repo query.

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
	if !strings.Contains(msg, "gh auth login") {
		t.Fatalf("expected gh auth login hint, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("expected empty stdout on auth error, got %q", resp.Stdout)
	}
	argv := readGhArgv(t, req.GhBin)
	if strings.Contains(argv, "repo list") {
		t.Fatalf("repo list should not run before auth, argv=%q", argv)
	}
}
```