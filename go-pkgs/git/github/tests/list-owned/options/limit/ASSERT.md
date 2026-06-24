## Expected

- `resp.GhArgv` contains `--limit` and `42`.

## Side Effects

- Mock `gh` invoked with limit flag.

## Errors

- `err` is nil.

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
	argv := resp.GhArgv
	if argv == "" {
		argv = readGhArgv(t, req.GhBin)
	}
	if !strings.Contains(argv, "--limit") {
		t.Fatalf("expected --limit in gh argv, got %q", argv)
	}
	if !strings.Contains(argv, "42") {
		t.Fatalf("expected limit 42 in gh argv, got %q", argv)
	}
}```