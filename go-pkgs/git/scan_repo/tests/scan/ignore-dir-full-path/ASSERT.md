## Expected

- `resp.Repos` is empty — full-path ignore skips the `scratch` tree.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 0 {
		t.Fatalf("expected 0 repos with full-path IgnoreDirs, got %d: %v", len(resp.Repos), resp.Repos)
	}
}
```