## Expected

- `GetTitle` returns a non-nil error when `ITERM_SESSION_ID` is empty.

## Errors

- Not-in-session error.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error when ITERM_SESSION_ID is empty")
	}
}
```
