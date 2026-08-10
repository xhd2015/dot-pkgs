## Expected

- `resp.Display` is `"child"`.

## Errors

- `err` is nil.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := "child"
	if resp.Display != want {
		t.Fatalf("expected %q, got %q for relative input %q", want, resp.Display, req.Path)
	}
}```
