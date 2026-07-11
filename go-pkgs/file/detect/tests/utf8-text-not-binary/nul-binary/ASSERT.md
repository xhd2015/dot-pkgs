## Expected

- `err` is nil.
- `resp.IsBinary == true`.
- `resp.Desc` is non-empty (typically `"binary file"`).

## Errors

- File with NUL classified as text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("DetectFileType(%q): %v", req.Path, err)
	}
	if !resp.IsBinary {
		t.Fatalf("DetectFileType(%q) isBinary=false desc=%q, want true for NUL content",
			req.Path, resp.Desc)
	}
	if resp.Desc == "" {
		t.Fatalf("DetectFileType(%q) desc empty, want binary description", req.Path)
	}
}
```
