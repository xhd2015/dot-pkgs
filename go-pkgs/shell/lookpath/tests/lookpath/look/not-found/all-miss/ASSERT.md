## Expected

- `err != nil`.
- Error message contains the binary name `"mytool"`.
- `resp.Path` is empty.

## Errors

- Not-found error from exhausted resolution pipeline.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertErrorContainsName(t, err, req.Name)
	if resp.Path != "" {
		t.Fatalf("Path = %q, want empty on not-found", resp.Path)
	}
}
```
