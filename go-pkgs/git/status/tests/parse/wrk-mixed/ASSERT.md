## Expected

- `resp.WrkCounts.Added` is 1 (`??` line).
- `resp.WrkCounts.Changed` is 1 (` M` line).
- `resp.WrkCounts.Renamed` is 1.
- `resp.WrkCounts.Deleted` is 1.

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
	if c.Added != 1 {
		t.Fatalf("Added = %d, want 1", c.Added)
	}
	if c.Changed != 1 {
		t.Fatalf("Changed = %d, want 1", c.Changed)
	}
	if c.Renamed != 1 {
		t.Fatalf("Renamed = %d, want 1", c.Renamed)
	}
	if c.Deleted != 1 {
		t.Fatalf("Deleted = %d, want 1", c.Deleted)
	}
}
```