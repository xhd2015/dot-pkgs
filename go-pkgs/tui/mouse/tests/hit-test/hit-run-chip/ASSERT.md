## Expected

- `err` is nil.
- `resp.HitOK` is true.
- `resp.Hit.ID` is `"run"`.

## Errors

- Miss or wrong chip ID when coordinates are inside the run rectangle.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("HitTest: %v", err)
	}
	if !resp.HitOK {
		t.Fatalf("HitTest(%d,%d) miss, want run chip", req.X, req.LocalY)
	}
	if resp.Hit.ID != "run" {
		t.Fatalf("HitTest ID=%q, want run", resp.Hit.ID)
	}
}
```
