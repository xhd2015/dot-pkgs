## Expected

- `resp.ExitCode` is 0.
- `resp.Stderr` is empty.
- `resp.Stdout` trimmed is `alice/zebra\towned`.
- Captured gh argv includes `repo list alice`.

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
	want := "alice/zebra\towned\n"
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch:\nwant %q\ngot  %q", want, resp.Stdout)
	}
	argv := readGhArgv(t, req.GhBin)
	if !strings.Contains(argv, "repo list alice") {
		t.Fatalf("expected repo list alice in argv, got %q", argv)
	}
}
```