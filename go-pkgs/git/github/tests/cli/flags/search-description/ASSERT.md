## Expected

- `resp.ExitCode` is 0.
- `resp.Stderr` is empty.
- `resp.Stdout` trimmed lines are exactly:
  - `alice/gadget\tdescription`
  - `alice/widget-lib\tdescription`
  in ascending sort order.
- Captured gh argv includes `search repos` and keyword `widget`.

## Side Effects

- Mock `gh api user` and `gh search repos` invoked; `repo list` not used.

## Errors

- Harness `err` is nil.

## Exit Code

- 0

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
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", resp.Stderr)
	}
	want := "alice/gadget\tdescription\nalice/widget-lib\tdescription\n"
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch:\nwant %q\ngot  %q", want, resp.Stdout)
	}
	argv := readGhArgv(t, req.GhBin)
	if !strings.Contains(argv, "search repos") {
		t.Fatalf("expected search repos in argv, got %q", argv)
	}
	if !strings.Contains(argv, "widget") {
		t.Fatalf("expected search keyword widget in argv, got %q", argv)
	}
}
```