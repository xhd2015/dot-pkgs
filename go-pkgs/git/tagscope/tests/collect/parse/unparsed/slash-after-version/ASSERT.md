## Expected

- `resp.ParseOK` is false.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	requireParseNotOK(t, resp.ParseOK)
}
```