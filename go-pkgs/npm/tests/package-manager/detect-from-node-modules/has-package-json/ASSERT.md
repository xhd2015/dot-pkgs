## Expected

- `resp.Trace.HasPackageJSON` is true.

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
	if !resp.Trace.HasPackageJSON {
		t.Fatal("expected HasPackageJSON = true")
	}
}```
