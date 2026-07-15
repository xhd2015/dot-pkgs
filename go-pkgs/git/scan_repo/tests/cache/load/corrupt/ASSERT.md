## Expected

- `resp` may be nil or unused when load fails.

## Errors

- `err` is non-nil (corrupt / invalid JSON).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error loading corrupt entry.json")
	}
}
```
