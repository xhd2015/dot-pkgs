## Expected

- `resp.GhArgv` contains `--no-archived`.

## Side Effects

- Mock `gh` invoked with archived filter.

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
	if !strings.Contains(argv, "--no-archived") {
		t.Fatalf("expected --no-archived in gh argv, got %q", argv)
	}
}```