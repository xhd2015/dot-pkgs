## Expected

- `err` is nil.
- `resp.IsBinary == false`.
- `resp.Desc` is not `"binary file"`.

## Errors

- ASCII text reported as binary.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("DetectFileType(%q): %v", req.Path, err)
	}
	if resp.IsBinary {
		t.Fatalf("DetectFileType(%q) isBinary=true desc=%q, want false for ASCII text",
			req.Path, resp.Desc)
	}
	if resp.Desc == "binary file" {
		t.Fatalf("DetectFileType(%q) desc=%q, want text-like", req.Path, resp.Desc)
	}
}
```
