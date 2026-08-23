## Expected

- `resp.WrkCounts.Staged` is 1.
- `resp.WrkCounts.Changed` is 0 (path-once; staged already covers the path).
- Other buckets are 0.

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
	c := resp.WrkCounts
	if c.Staged != 1 {
		t.Fatalf("Staged = %d, want 1", c.Staged)
	}
	if c.Changed != 0 {
		t.Fatalf("Changed = %d, want 0", c.Changed)
	}
	if c.Renamed != 0 || c.Deleted != 0 || c.Untracked != 0 {
		t.Fatalf("unexpected buckets: %+v", c)
	}
}
```
