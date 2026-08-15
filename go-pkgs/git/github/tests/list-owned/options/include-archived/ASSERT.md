## Expected

- `resp.GhArgv` does NOT contain `--no-archived`.

## Side Effects

- Mock `gh` invoked without archived exclusion flag.

## Errors

- `err` is nil.

## Exit Code

- N/A (library call).

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
	argv := resp.GhArgv
	if argv == "" {
		argv = readGhArgv(t, req.GhBin)
	}
	if strings.Contains(argv, "--no-archived") {
		t.Fatalf("did not expect --no-archived in gh argv, got %q", argv)
	}
}```