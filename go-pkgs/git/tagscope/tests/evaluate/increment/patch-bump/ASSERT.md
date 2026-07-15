## Expected

- `NextTag` is `v0.0.10`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.NextTag != "v0.0.10" {
		t.Fatalf("NextTag = %q, want v0.0.10", resp.NextTag)
	}
}
```