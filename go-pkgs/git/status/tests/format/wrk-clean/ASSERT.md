## Expected

- `resp.Formatted` is `"clean"`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Formatted != "clean" {
		t.Fatalf("formatted = %q, want clean", resp.Formatted)
	}
}
```