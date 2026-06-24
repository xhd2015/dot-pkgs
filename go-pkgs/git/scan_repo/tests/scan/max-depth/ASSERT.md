## Expected

- `resp.Repos` is empty — repo at depth 3 exceeds `MaxDepth=2`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 0 {
		t.Fatalf("expected 0 repos with MaxDepth=2, got %d: %v", len(resp.Repos), resp.Repos)
	}
}
```