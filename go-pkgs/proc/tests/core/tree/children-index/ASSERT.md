## Expected

- `resp.Children[10]` is `[20, 30]` (sorted ascending).
- `resp.Children[20]` is `[40]`.
- `resp.Children[1]` is `[10]`.
- Keys without children may be absent (not required present as empty slices).

## Errors

- `err` is nil.
- Unsorted sibling PIDs is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp.Children == nil {
		t.Fatal("Children is nil")
	}
	assertIntSliceEqual(t, resp.Children[10], []int{20, 30})
	assertIntSliceEqual(t, resp.Children[20], []int{40})
	assertIntSliceEqual(t, resp.Children[1], []int{10})
}
```
