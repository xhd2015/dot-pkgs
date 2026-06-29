## Expected

- `resp.Repos` is empty — `scratch` basename is ignored via `IgnoreDirBasenames`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Repos) != 0 {
		t.Fatalf("expected 0 repos with IgnoreDirBasenames scratch, got %d: %v", len(resp.Repos), resp.Repos)
	}
}
```