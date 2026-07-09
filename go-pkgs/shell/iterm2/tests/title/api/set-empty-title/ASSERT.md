## Expected

- `SetTitle` returns a non-nil error for empty title (no successful set).

## Errors

- Empty title validation error.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}
```
