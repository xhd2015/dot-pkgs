## Expected

- `resp.Counts.Modified` is 2.
- `resp.Counts.Untracked` is 1.
- Other count fields are zero.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	c := resp.Counts
	if c.Modified != 2 {
		t.Fatalf("Modified = %d, want 2", c.Modified)
	}
	if c.Untracked != 1 {
		t.Fatalf("Untracked = %d, want 1", c.Untracked)
	}
	if c.Added != 0 || c.Deleted != 0 || c.Renamed != 0 || c.Copied != 0 || c.Unmerged != 0 {
		t.Fatalf("unexpected non-zero counts: %+v", c)
	}
}
```
