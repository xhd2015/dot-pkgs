## Expected

- `err` is nil.
- `resp.Replies` is empty (status DSR is not CPR request).
- `resp.Rest` is empty.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Replies) != 0 {
		t.Fatalf("should not reply CPR to ESC[5n: %q", resp.Replies)
	}
	if len(resp.Rest) != 0 {
		t.Fatalf("rest should be empty after complete non-query, got %q", resp.Rest)
	}
}
```
