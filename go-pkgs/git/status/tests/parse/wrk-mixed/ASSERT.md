## Expected

- `resp.WrkCounts.Changed` is 1 (` M` line).
- `resp.WrkCounts.Untracked` is 1 (`??` line).
- `resp.WrkCounts.Staged` is 1 (`R  ` staged rename).
- `resp.WrkCounts.Deleted` is 1 (` D` unstaged delete).
- `resp.WrkCounts.Renamed` is 0 (staged rename is not also renamed).

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
	if c.Changed != 1 {
		t.Fatalf("Changed = %d, want 1", c.Changed)
	}
	if c.Renamed != 0 {
		t.Fatalf("Renamed = %d, want 0", c.Renamed)
	}
	if c.Deleted != 1 {
		t.Fatalf("Deleted = %d, want 1", c.Deleted)
	}
	if c.Untracked != 1 {
		t.Fatalf("Untracked = %d, want 1", c.Untracked)
	}
}
```
