## Expected

- All count fields in `resp.Counts` are zero.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	c := resp.Counts
	if c.Modified != 0 || c.Added != 0 || c.Deleted != 0 || c.Untracked != 0 ||
		c.Renamed != 0 || c.Copied != 0 || c.Unmerged != 0 {
		t.Fatalf("expected zero counts, got %+v", c)
	}
}
```
