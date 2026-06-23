## Expected

- `resp.Display` equals `req.Path` unchanged (`"/abs/path"`).

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Display != req.Path {
		t.Fatalf("expected unchanged %q, got %q", req.Path, resp.Display)
	}
}```
