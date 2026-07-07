## Expected

- `resp.Output` is `"true"`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Output != "true" {
		t.Fatalf("output = %q, want true", resp.Output)
	}
}
```
