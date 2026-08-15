## Expected

- `resp.Repos` is non-nil empty slice (`len` 0).

## Side Effects

- Mock `gh` invoked once.

## Errors

- `err` is nil.

## Exit Code

- N/A (library call).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Repos == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(resp.Repos) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(resp.Repos))
	}
}```