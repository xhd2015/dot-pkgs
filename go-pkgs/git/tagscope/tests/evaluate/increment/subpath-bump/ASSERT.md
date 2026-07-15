## Expected

- `NextTag` is `sub/v0.2.10`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.NextTag != "sub/v0.2.10" {
		t.Fatalf("NextTag = %q, want sub/v0.2.10", resp.NextTag)
	}
}
```