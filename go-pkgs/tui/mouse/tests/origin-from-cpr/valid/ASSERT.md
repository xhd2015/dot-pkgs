## Expected

- `err` is nil.
- `resp.OriginOK` is true.
- `resp.OriginY0` is 6.

## Errors

- Rejecting a valid mid-pane CPR or returning wrong origin.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("OriginFromCPR: %v", err)
	}
	if !resp.OriginOK || resp.OriginY0 != 6 {
		t.Fatalf("OriginFromCPR(%d,%d)=(%d,%v), want (6,true)",
			req.Row1, req.ViewLines, resp.OriginY0, resp.OriginOK)
	}
}
```
