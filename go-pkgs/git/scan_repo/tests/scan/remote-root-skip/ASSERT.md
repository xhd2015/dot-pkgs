## Expected

- `err` is nil.
- `Repos` is empty.

## Errors

- No error returned from `Run`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 0 {
		t.Fatalf("expected 0 repos for remote-backed root, got %d: %v", len(resp.Repos), resp.Repos)
	}
}
```